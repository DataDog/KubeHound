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
	"github.com/DataDog/KubeHound/pkg/globals"
	"github.com/DataDog/KubeHound/pkg/globals/types"
	"github.com/DataDog/KubeHound/pkg/telemetry"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	"github.com/DataDog/KubeHound/pkg/telemetry/statsd"
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
	nodePath                = "nodes.json"
	endpointPath            = "endpointslices.discovery.k8s.io.json"
	clusterRolesPath        = "clusterroles.rbac.authorization.k8s.io.json"
	clusterRoleBindingsPath = "clusterrolebindings.rbac.authorization.k8s.io.json"
	podPath                 = "pods.json"
	rolesPath               = "roles.rbac.authorization.k8s.io.json"
	roleBindingsPath        = "rolebindings.rbac.authorization.k8s.io.json"
)

const (
	FileCollectorName = "local-file-collector"
)

// FileCollector implements a collector based on local K8s API json files generated outside the KubeHound application via e.g kubectl.
type FileCollector struct {
	cfg  *config.FileCollectorConfig
	log  *log.KubehoundLogger
	tags []string
}

// NewFileCollector creates a new instance of the file collector from the provided application config.
func NewFileCollector(ctx context.Context, cfg *config.KubehoundConfig) (CollectorClient, error) {
	baseTags := append(telemetry.BaseTags, telemetry.TagCollectorTypeFile)
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
		tags: baseTags,
	}, nil
}

func (c *FileCollector) Name() string {
	return FileCollectorName
}

func (c *FileCollector) HealthCheck(ctx context.Context) (bool, error) {
	file, err := os.Stat(c.cfg.Directory)
	if err != nil {
		return false, fmt.Errorf("file collector base path: %w", err)
	}

	if !file.IsDir() {
		return false, fmt.Errorf("file collector base path is not a directory")
	}

	return true, nil
}

func (c *FileCollector) Close(ctx context.Context) error {
	// NOP for this implementation
	return nil
}

func (c *FileCollector) Cluster(ctx context.Context) string {
	return "TODO"
}

// streamPodsNamespace streams the pod objects in a single file, corresponding to a cluster namespace.
func (c *FileCollector) streamPodsNamespace(ctx context.Context, fp string, ingestor PodIngestor) error {
	list, err := readList[corev1.PodList](ctx, fp)
	if err != nil {
		return err
	}

	for _, item := range list.Items {
		_ = statsd.Incr(telemetry.MetricCollectorPodsCount, c.tags, 1)
		i := types.PodType(&item)
		err = ingestor.IngestPod(ctx, i)
		if err != nil {
			return fmt.Errorf("processing K8s pod %s: %w", i.Name, err)
		}
	}

	return nil
}

func (c *FileCollector) StreamPods(ctx context.Context, ingestor PodIngestor) error {
	span, ctx := tracer.StartSpanFromContext(ctx, telemetry.SpanOperationStream, tracer.Measured())
	span.SetTag(telemetry.TagKeyResource, telemetry.TagResourcePods)
	defer span.Finish()

	err := filepath.WalkDir(c.cfg.Directory, func(path string, d fs.DirEntry, err error) error {
		if path == c.cfg.Directory || !d.IsDir() {
			// Skip files
			return nil
		}

		fp := filepath.Join(path, podPath)
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
		_ = statsd.Incr(telemetry.MetricCollectorRolesCount, c.tags, 1)
		i := types.RoleType(&item)
		err = ingestor.IngestRole(ctx, i)
		if err != nil {
			return fmt.Errorf("processing K8s role %s: %w", i.Name, err)
		}
	}

	return nil
}

func (c *FileCollector) StreamRoles(ctx context.Context, ingestor RoleIngestor) error {
	span, ctx := tracer.StartSpanFromContext(ctx, telemetry.SpanOperationStream, tracer.Measured())
	span.SetTag(telemetry.TagKeyResource, telemetry.TagResourceRoles)
	defer span.Finish()

	err := filepath.WalkDir(c.cfg.Directory, func(path string, d fs.DirEntry, err error) error {
		if path == c.cfg.Directory || !d.IsDir() {
			// Skip files
			return nil
		}

		f := filepath.Join(path, rolesPath)
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
		_ = statsd.Incr(telemetry.MetricCollectorRoleBindingsCount, c.tags, 1)
		i := types.RoleBindingType(&item)
		err = ingestor.IngestRoleBinding(ctx, i)
		if err != nil {
			return fmt.Errorf("processing K8s role binding %s: %w", i.Name, err)
		}
	}

	return nil
}

func (c *FileCollector) StreamRoleBindings(ctx context.Context, ingestor RoleBindingIngestor) error {
	span, ctx := tracer.StartSpanFromContext(ctx, telemetry.SpanOperationStream, tracer.Measured())
	span.SetTag(telemetry.TagKeyResource, telemetry.TagResourceRolebindings)
	defer span.Finish()

	err := filepath.WalkDir(c.cfg.Directory, func(path string, d fs.DirEntry, err error) error {
		if path == c.cfg.Directory || !d.IsDir() {
			// Skip files
			return nil
		}

		fp := filepath.Join(path, roleBindingsPath)
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
		_ = statsd.Incr(telemetry.MetricCollectorEndpointCount, c.tags, 1)
		i := types.EndpointType(&item)
		err = ingestor.IngestEndpoint(ctx, i)
		if err != nil {
			return fmt.Errorf("processing K8s endpoint slice %s: %w", i.Name, err)
		}
	}

	return nil
}

func (c *FileCollector) StreamEndpoints(ctx context.Context, ingestor EndpointIngestor) error {
	span, ctx := tracer.StartSpanFromContext(ctx, telemetry.SpanOperationStream, tracer.Measured())
	span.SetTag(telemetry.TagKeyResource, telemetry.TagResourceEndpoints)
	defer span.Finish()

	err := filepath.WalkDir(c.cfg.Directory, func(path string, d fs.DirEntry, err error) error {
		if path == c.cfg.Directory || !d.IsDir() {
			// Skip files
			return nil
		}

		fp := filepath.Join(path, endpointPath)
		c.log.Debugf("Streaming endpoint slices from file %s", fp)

		return c.streamEndpointsNamespace(ctx, fp, ingestor)
	})

	if err != nil {
		return fmt.Errorf("file collector stream endpoint slices: %w", err)
	}

	return ingestor.Complete(ctx)
}

func (c *FileCollector) StreamNodes(ctx context.Context, ingestor NodeIngestor) error {
	span, ctx := tracer.StartSpanFromContext(ctx, telemetry.SpanOperationStream, tracer.Measured())
	span.SetTag(telemetry.TagKeyResource, telemetry.TagResourceNodes)
	defer span.Finish()

	fp := filepath.Join(c.cfg.Directory, nodePath)
	c.log.Debugf("Streaming nodes from file %s", fp)

	list, err := readList[corev1.NodeList](ctx, fp)
	if err != nil {
		return err
	}

	for _, item := range list.Items {
		_ = statsd.Incr(telemetry.MetricCollectorNodesCount, c.tags, 1)
		i := types.NodeType(&item)
		err = ingestor.IngestNode(ctx, i)
		if err != nil {
			return fmt.Errorf("processing K8s node %s::%s: %w", i.Namespace, i.Name, err)
		}
	}

	return ingestor.Complete(ctx)
}

func (c *FileCollector) StreamClusterRoles(ctx context.Context, ingestor ClusterRoleIngestor) error {
	span, ctx := tracer.StartSpanFromContext(ctx, telemetry.SpanOperationStream, tracer.Measured())
	span.SetTag(telemetry.TagKeyResource, telemetry.TagResourceClusterRoles)
	defer span.Finish()

	fp := filepath.Join(c.cfg.Directory, clusterRolesPath)
	c.log.Debugf("Streaming cluster roles from file %s", fp)

	list, err := readList[rbacv1.ClusterRoleList](ctx, fp)
	if err != nil {
		return err
	}

	for _, item := range list.Items {
		_ = statsd.Incr(telemetry.MetricCollectorClusterRolesCount, c.tags, 1)
		i := types.ClusterRoleType(&item)
		err = ingestor.IngestClusterRole(ctx, i)
		if err != nil {
			return fmt.Errorf("processing k8s cluster role %s: %w", i.Name, err)
		}
	}

	return ingestor.Complete(ctx)
}

func (c *FileCollector) StreamClusterRoleBindings(ctx context.Context, ingestor ClusterRoleBindingIngestor) error {
	span, ctx := tracer.StartSpanFromContext(ctx, telemetry.SpanOperationStream, tracer.Measured())
	span.SetTag(telemetry.TagKeyResource, telemetry.TagResourceClusterRolebindings)
	defer span.Finish()

	fp := filepath.Join(c.cfg.Directory, clusterRoleBindingsPath)
	c.log.Debugf("Streaming cluster role bindings from file %s", fp)

	list, err := readList[rbacv1.ClusterRoleBindingList](ctx, fp)
	if err != nil {
		return err
	}

	for _, item := range list.Items {
		_ = statsd.Incr(telemetry.MetricCollectorClusterRoleBindingsCount, c.tags, 1)
		i := types.ClusterRoleBindingType(&item)
		err = ingestor.IngestClusterRoleBinding(ctx, i)
		if err != nil {
			return fmt.Errorf("processing K8s cluster role binding %s: %w", i.Name, err)
		}
	}

	return ingestor.Complete(ctx)
}

// readList loads a list of K8s API objects into memory from a JSON file on disk.
// NOTE: This implementation reads the entire array of objects from the file into memory at once.
func readList[Tl types.ListInputType](ctx context.Context, inputPath string) (Tl, error) {
	span, _ := tracer.StartSpanFromContext(ctx, telemetry.SpanOperationReadFile, tracer.Measured())
	defer span.Finish()

	var inputList Tl
	bytes, err := os.ReadFile(inputPath)
	if err != nil {
		return inputList, fmt.Errorf("read file %s: %v", inputPath, err)
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
