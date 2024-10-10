package collector

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/DataDog/KubeHound/pkg/config"
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
	MetadataPath            = "metadata.json"
)

const (
	FileCollectorName = "local-file-collector"
)

// FileCollector implements a collector based on local K8s API json files generated outside the KubeHound application via e.g kubectl.
type FileCollector struct {
	cfg         *config.FileCollectorConfig
	log         *log.KubehoundLogger
	tags        collectorTags
	clusterName string
}

// NewFileCollector creates a new instance of the file collector from the provided application config.
func NewFileCollector(ctx context.Context, cfg *config.KubehoundConfig) (CollectorClient, error) {
	if cfg.Collector.Type != config.CollectorTypeFile {
		return nil, fmt.Errorf("invalid collector type in config: %s", cfg.Collector.Type)
	}

	if cfg.Collector.File == nil {
		return nil, errors.New("file collector config not provided")
	}

	l := log.Trace(ctx)
	l.Info("Creating file collector from directory", log.String("path", cfg.Collector.File.Directory))

	return &FileCollector{
		cfg: cfg.Collector.File,
		// log:         l,
		tags:        newCollectorTags(),
		clusterName: cfg.Dynamic.ClusterName,
	}, nil
}

// This function has no meaning in the file collector as it should already have all the metadata gathered in the dumped files.
func (c *FileCollector) ComputeMetadata(ctx context.Context, ingestor MetadataIngestor) error {
	return nil
}

func (c *FileCollector) Name() string {
	return FileCollectorName
}

func (c *FileCollector) HealthCheck(_ context.Context) (bool, error) {
	file, err := os.Stat(c.cfg.Directory)
	if err != nil {
		return false, fmt.Errorf("file collector base path: %s %w", c.cfg.Directory, err)
	}

	if !file.IsDir() {
		return false, fmt.Errorf("file collector base path is not a directory: %s", file.Name())
	}

	if c.clusterName == "" {
		return false, errors.New("file collector cluster name not provided")
	}

	return true, nil
}

func (c *FileCollector) ClusterInfo(ctx context.Context) (*config.ClusterInfo, error) {
	return &config.ClusterInfo{
		Name: c.clusterName,
	}, nil
}

func (c *FileCollector) Close(_ context.Context) error {
	// NOP for this implementation
	return nil
}

// streamPodsNamespace streams the pod objects in a single file, corresponding to a cluster namespace.
func (c *FileCollector) streamPodsNamespace(ctx context.Context, fp string, ingestor PodIngestor) error {
	list, err := readList[corev1.PodList](ctx, fp)
	if err != nil {
		return err
	}

	for _, item := range list.Items {
		_ = statsd.Incr(metric.CollectorCount, c.tags.pod, 1)
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
	var err error
	defer func() { span.Finish(tracer.WithError(err)) }()

	err = filepath.WalkDir(c.cfg.Directory, func(path string, d fs.DirEntry, err error) error {
		if path == c.cfg.Directory || !d.IsDir() {
			// Skip files
			return nil
		}

		fp := filepath.Join(path, PodPath)

		// Check if the file exists
		if _, err := os.Stat(fp); os.IsNotExist(err) {
			// Skipping streaming as file does not exist (k8s type not necessary required in a namespace, for instance, an namespace can have no pods)
			return nil
		}

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
		_ = statsd.Incr(metric.CollectorCount, c.tags.role, 1)
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
	var err error
	defer func() { span.Finish(tracer.WithError(err)) }()

	err = filepath.WalkDir(c.cfg.Directory, func(path string, d fs.DirEntry, err error) error {
		if path == c.cfg.Directory || !d.IsDir() {
			// Skip files
			return nil
		}

		f := filepath.Join(path, RolesPath)

		// Check if the file exists
		if _, err := os.Stat(f); os.IsNotExist(err) {
			// Skipping streaming as file does not exist (k8s type not necessary required in a namespace, for instance, an namespace can have no roles)
			return nil
		}

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
		_ = statsd.Incr(metric.CollectorCount, c.tags.rolebinding, 1)
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
	var err error
	defer func() { span.Finish(tracer.WithError(err)) }()

	err = filepath.WalkDir(c.cfg.Directory, func(path string, d fs.DirEntry, err error) error {
		if path == c.cfg.Directory || !d.IsDir() {
			// Skip files
			return nil
		}

		fp := filepath.Join(path, RoleBindingsPath)

		// Check if the file exists
		if _, err := os.Stat(fp); os.IsNotExist(err) {
			// Skipping streaming as file does not exist (k8s type not necessary required in a namespace, for instance, an namespace can have no rolebindings)
			return nil
		}

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
		_ = statsd.Incr(metric.CollectorCount, c.tags.endpoint, 1)
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
	var err error
	defer func() { span.Finish(tracer.WithError(err)) }()

	err = filepath.WalkDir(c.cfg.Directory, func(path string, d fs.DirEntry, err error) error {
		if path == c.cfg.Directory || !d.IsDir() {
			// Skip files
			return nil
		}

		fp := filepath.Join(path, EndpointPath)

		// Check if the file exists
		if _, err := os.Stat(fp); os.IsNotExist(err) {
			// Skipping streaming as file does not exist (k8s type not necessary required in a namespace, for instance, an namespace can have no endpoints)
			return nil
		}

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
	var err error
	defer func() { span.Finish(tracer.WithError(err)) }()

	fp := filepath.Join(c.cfg.Directory, NodePath)
	c.log.Debugf("Streaming nodes from file %s", fp)

	list, err := readList[corev1.NodeList](ctx, fp)
	if err != nil {
		return err
	}

	for _, item := range list.Items {
		_ = statsd.Incr(metric.CollectorCount, c.tags.node, 1)
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
	var err error
	defer func() { span.Finish(tracer.WithError(err)) }()

	fp := filepath.Join(c.cfg.Directory, ClusterRolesPath)
	c.log.Debugf("Streaming cluster roles from file %s", fp)

	list, err := readList[rbacv1.ClusterRoleList](ctx, fp)
	if err != nil {
		return err
	}

	for _, item := range list.Items {
		_ = statsd.Incr(metric.CollectorCount, c.tags.clusterrole, 1)
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
	var err error
	defer func() { span.Finish(tracer.WithError(err)) }()

	fp := filepath.Join(c.cfg.Directory, ClusterRoleBindingsPath)
	c.log.Debugf("Streaming cluster role bindings from file %s", fp)

	list, err := readList[rbacv1.ClusterRoleBindingList](ctx, fp)
	if err != nil {
		return err
	}

	for _, item := range list.Items {
		_ = statsd.Incr(metric.CollectorCount, c.tags.clusterrolebinding, 1)
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
	var err error
	defer func() { span.Finish(tracer.WithError(err)) }()

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
		return inputList, fmt.Errorf("unmarshalling %T in %s json: %w", inputList, inputPath, err)
	}

	return inputList, nil
}
