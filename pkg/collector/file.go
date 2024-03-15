package collector

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync"

	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/DataDog/KubeHound/pkg/globals"
	"github.com/DataDog/KubeHound/pkg/globals/types"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	"github.com/DataDog/KubeHound/pkg/telemetry/metric"
	"github.com/DataDog/KubeHound/pkg/telemetry/span"
	"github.com/DataDog/KubeHound/pkg/telemetry/statsd"
	"github.com/DataDog/KubeHound/pkg/telemetry/tag"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	rbacv1 "k8s.io/api/rbac/v1"
)

// Expect a file structure of the following
// |____<namespace>
// | |____rolebindings.rbac.authorization.k8s.io.json
// | |____pods.json
// | |____endpointslices.discovery.k8s.io.json
// | |____roles.rbac.authorization.k8s.io.json
// |____<namespace>
// | |____rolebindings.rbac.authorization.k8s.io.json
// | |____pods.json
// | |____endpointslices.discovery.k8s.io.json
// | |____roles.rbac.authorization.k8s.io.json
// |____nodes.json
// |____clusterroles.rbac.authorization.k8s.io.json
// |____clusterrolebindings.rbac.authorization.k8s.io.json
const (
	NodePath                = "nodes.json"
	EndpointPath            = "endpointslices.discovery.k8s.io.json"
	ClusterRolesPath        = "clusterroles.rbac.authorization.k8s.io.json"
	ClusterRoleBindingsPath = "clusterrolebindings.rbac.authorization.k8s.io.json"
	PodPath                 = "pods.json"
	RolesPath               = "roles.rbac.authorization.k8s.io.json"
	RoleBindingsPath        = "rolebindings.rbac.authorization.k8s.io.json"
)

const (
	FileCollectorName = "local-file-collector"
)

// FileCollector implements a collector based on local K8s API json files generated outside the KubeHound application via e.g kubectl.
type FileCollector struct {
	cfg  *config.FileCollectorConfig
	log  *log.KubehoundLogger
	tags []string
	mu   sync.Mutex
}

// NewFileCollector creates a new instance of the file collector from the provided application config.
func NewFileCollector(ctx context.Context, cfg *config.KubehoundConfig) (CollectorClient, error) {
	tags := tag.BaseTags
	tags = append(tags, tag.Collector(FileCollectorName))
	if cfg.Collector.Type != config.CollectorTypeFile {
		return nil, fmt.Errorf("invalid collector type in config: %s", cfg.Collector.Type)
	}

	if cfg.Collector.File == nil {
		return nil, errors.New("file collector config not provided")
	}

	l := log.Trace(ctx, log.WithComponent(globals.FileCollectorComponent))
	l.Infof("Creating file collector from directory %s", cfg.Collector.File.Directory)

	return &FileCollector{
		cfg:  cfg.Collector.File,
		log:  l,
		tags: tag.GetBaseTagsWith(tag.Collector(FileCollectorName)),
		mu:   sync.Mutex{},
	}, nil
}

func (c *FileCollector) Name() string {
	return FileCollectorName
}

func (c *FileCollector) HealthCheck(_ context.Context) (bool, error) {
	file, err := os.Stat(c.cfg.Directory)
	if err != nil {
		return false, fmt.Errorf("file collector base path: %w", err)
	}

	if !file.IsDir() {
		return false, fmt.Errorf("file collector base path is not a directory")
	}

	if c.cfg.ClusterName == "" {
		return false, errors.New("file collector cluster name not provided")
	}

	return true, nil
}

func (c *FileCollector) ClusterInfo(ctx context.Context) (*ClusterInfo, error) {
	return &ClusterInfo{
		Name: c.cfg.ClusterName,
	}, nil
}

func (c *FileCollector) Tags(ctx context.Context) []string {
	return c.tags
}

func (c *FileCollector) Close(_ context.Context) error {
	// NOP for this implementation
	return nil
}

func (c *FileCollector) count(entity string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.tags = append(c.tags, tag.Entity(entity))
	_ = statsd.Incr(metric.CollectorCount, c.tags, 1)
}

// streamPodsNamespace streams the pod objects in a single file, corresponding to a cluster namespace.
func (c *FileCollector) streamPodsNamespace(ctx context.Context, fp string, ingestor PodIngestor) error {
	list, err := readList[corev1.PodList](ctx, fp)
	if err != nil {
		return err
	}

	for _, item := range list.Items {
		c.count(tag.EntityPods)
		i := item
		err = ingestor.IngestPod(ctx, &i)
		if err != nil {
			return fmt.Errorf("processing K8s pod %s: %w", i.Name, err)
		}
	}

	return nil
}

func (c *FileCollector) StreamPods(ctx context.Context, ingestor PodIngestor) error {
	span, ctx := tracer.StartSpanFromContext(ctx, span.CollectorStream, tracer.Measured())
	span.SetTag(tag.EntityTag, tag.EntityPods)
	defer span.Finish()

	err := filepath.WalkDir(c.cfg.Directory, func(path string, d fs.DirEntry, err error) error {
		if path == c.cfg.Directory || !d.IsDir() {
			// Skip files
			return nil
		}

		fp := filepath.Join(path, PodPath)
		c.log.Debugf("Streaming pods from file %s", fp)

		return c.streamPodsNamespace(ctx, fp, ingestor)
	})

	if err != nil {
		return fmt.Errorf("file collector stream pods: %w", err)
	}

	return ingestor.Complete(ctx)
}

// streamRolesNamespace streams the role objects in a single file, corresponding to a cluster namespace.
func (c *FileCollector) streamRolesNamespace(ctx context.Context, fp string, ingestor RoleIngestor) error {
	list, err := readList[rbacv1.RoleList](ctx, fp)
	if err != nil {
		return err
	}

	for _, item := range list.Items {
		c.count(tag.EntityRoles)
		i := item
		err = ingestor.IngestRole(ctx, &i)
		if err != nil {
			return fmt.Errorf("processing K8s role %s: %w", i.Name, err)
		}
	}

	return nil
}

func (c *FileCollector) StreamRoles(ctx context.Context, ingestor RoleIngestor) error {
	span, ctx := tracer.StartSpanFromContext(ctx, span.CollectorStream, tracer.Measured())
	span.SetTag(tag.EntityTag, tag.EntityRoles)
	defer span.Finish()

	err := filepath.WalkDir(c.cfg.Directory, func(path string, d fs.DirEntry, err error) error {
		if path == c.cfg.Directory || !d.IsDir() {
			// Skip files
			return nil
		}

		f := filepath.Join(path, RolesPath)
		c.log.Debugf("Streaming roles from file %s", f)

		return c.streamRolesNamespace(ctx, f, ingestor)
	})

	if err != nil {
		return fmt.Errorf("file collector stream roles: %w", err)
	}

	return ingestor.Complete(ctx)
}

// streamRoleBindingsNamespace streams the role bindings objects in a single file, corresponding to a cluster namespace.
func (c *FileCollector) streamRoleBindingsNamespace(ctx context.Context, fp string, ingestor RoleBindingIngestor) error {
	list, err := readList[rbacv1.RoleBindingList](ctx, fp)
	if err != nil {
		return err
	}

	for _, item := range list.Items {
		c.count(tag.EntityRolebindings)
		i := item
		err = ingestor.IngestRoleBinding(ctx, &i)
		if err != nil {
			return fmt.Errorf("processing K8s role binding %s: %w", i.Name, err)
		}
	}

	return nil
}

func (c *FileCollector) StreamRoleBindings(ctx context.Context, ingestor RoleBindingIngestor) error {
	span, ctx := tracer.StartSpanFromContext(ctx, span.CollectorStream, tracer.Measured())
	span.SetTag(tag.EntityTag, tag.EntityRolebindings)
	defer span.Finish()

	err := filepath.WalkDir(c.cfg.Directory, func(path string, d fs.DirEntry, err error) error {
		if path == c.cfg.Directory || !d.IsDir() {
			// Skip files
			return nil
		}

		fp := filepath.Join(path, RoleBindingsPath)
		c.log.Debugf("Streaming role bindings from file %s", fp)

		return c.streamRoleBindingsNamespace(ctx, fp, ingestor)
	})

	if err != nil {
		return fmt.Errorf("file collector stream role bindings: %w", err)
	}

	return ingestor.Complete(ctx)
}

// streamEndpointsNamespace streams the endpoint slices in a single file, corresponding to a cluster namespace.
func (c *FileCollector) streamEndpointsNamespace(ctx context.Context, fp string, ingestor EndpointIngestor) error {
	list, err := readList[discoveryv1.EndpointSliceList](ctx, fp)
	if err != nil {
		return err
	}

	for _, item := range list.Items {
		c.count(tag.EntityEndpoints)
		i := item
		err = ingestor.IngestEndpoint(ctx, &i)
		if err != nil {
			return fmt.Errorf("processing K8s endpoint slice %s: %w", i.Name, err)
		}
	}

	return nil
}

func (c *FileCollector) StreamEndpoints(ctx context.Context, ingestor EndpointIngestor) error {
	span, ctx := tracer.StartSpanFromContext(ctx, span.CollectorStream, tracer.Measured())
	span.SetTag(tag.EntityTag, tag.EntityEndpoints)
	defer span.Finish()

	err := filepath.WalkDir(c.cfg.Directory, func(path string, d fs.DirEntry, err error) error {
		if path == c.cfg.Directory || !d.IsDir() {
			// Skip files
			return nil
		}

		fp := filepath.Join(path, EndpointPath)
		c.log.Debugf("Streaming endpoint slices from file %s", fp)

		return c.streamEndpointsNamespace(ctx, fp, ingestor)
	})

	if err != nil {
		return fmt.Errorf("file collector stream endpoint slices: %w", err)
	}

	return ingestor.Complete(ctx)
}

func (c *FileCollector) StreamNodes(ctx context.Context, ingestor NodeIngestor) error {
	span, ctx := tracer.StartSpanFromContext(ctx, span.CollectorStream, tracer.Measured())
	span.SetTag(tag.EntityTag, tag.EntityNodes)
	defer span.Finish()

	fp := filepath.Join(c.cfg.Directory, NodePath)
	c.log.Debugf("Streaming nodes from file %s", fp)

	list, err := readList[corev1.NodeList](ctx, fp)
	if err != nil {
		return err
	}

	for _, item := range list.Items {
		c.count(tag.EntityNodes)
		i := item
		err = ingestor.IngestNode(ctx, &i)
		if err != nil {
			return fmt.Errorf("processing K8s node %s::%s: %w", i.Namespace, i.Name, err)
		}
	}

	return ingestor.Complete(ctx)
}

func (c *FileCollector) StreamClusterRoles(ctx context.Context, ingestor ClusterRoleIngestor) error {
	span, ctx := tracer.StartSpanFromContext(ctx, span.CollectorStream, tracer.Measured())
	span.SetTag(tag.EntityTag, tag.EntityClusterRoles)
	defer span.Finish()

	fp := filepath.Join(c.cfg.Directory, ClusterRolesPath)
	c.log.Debugf("Streaming cluster roles from file %s", fp)

	list, err := readList[rbacv1.ClusterRoleList](ctx, fp)
	if err != nil {
		return err
	}

	for _, item := range list.Items {
		c.count(tag.EntityClusterRoles)
		i := item
		err = ingestor.IngestClusterRole(ctx, &i)
		if err != nil {
			return fmt.Errorf("processing k8s cluster role %s: %w", i.Name, err)
		}
	}

	return ingestor.Complete(ctx)
}

func (c *FileCollector) StreamClusterRoleBindings(ctx context.Context, ingestor ClusterRoleBindingIngestor) error {
	span, ctx := tracer.StartSpanFromContext(ctx, span.CollectorStream, tracer.Measured())
	span.SetTag(tag.EntityTag, tag.EntityClusterRolebindings)
	defer span.Finish()

	fp := filepath.Join(c.cfg.Directory, ClusterRoleBindingsPath)
	c.log.Debugf("Streaming cluster role bindings from file %s", fp)

	list, err := readList[rbacv1.ClusterRoleBindingList](ctx, fp)
	if err != nil {
		return err
	}

	for _, item := range list.Items {
		_ = statsd.Incr(metric.CollectorCount, append(c.tags, tag.Entity(tag.EntityClusterRolebindings)), 1)
		i := item
		err = ingestor.IngestClusterRoleBinding(ctx, &i)
		if err != nil {
			return fmt.Errorf("processing K8s cluster role binding %s: %w", i.Name, err)
		}
	}

	return ingestor.Complete(ctx)
}

// readList loads a list of K8s API objects into memory from a JSON file on disk.
// NOTE: This implementation reads the entire array of objects from the file into memory at once.
func readList[Tl types.ListInputType](ctx context.Context, inputPath string) (Tl, error) {
	span, _ := tracer.StartSpanFromContext(ctx, span.DumperReadFile, tracer.Measured())
	defer span.Finish()

	var inputList Tl
	bytes, err := os.ReadFile(inputPath)
	if err != nil {
		return inputList, fmt.Errorf("read file %s: %w", inputPath, err)
	}

	if len(bytes) == 0 {
		return inputList, nil
	}

	err = json.Unmarshal(bytes, &inputList)
	if err != nil {
		return inputList, fmt.Errorf("unmarshalling %T json: %w", inputList, err)
	}

	return inputList, nil
}
