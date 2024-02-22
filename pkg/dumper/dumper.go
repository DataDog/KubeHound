package dumper

import (
	"context"
	"encoding/json"
	"fmt"
	"path"

	"time"

	"github.com/DataDog/KubeHound/pkg/collector"
	"github.com/DataDog/KubeHound/pkg/dumper/writer"
	"github.com/DataDog/KubeHound/pkg/globals/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/ingestor/preflight"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	"github.com/DataDog/KubeHound/pkg/telemetry/metric"
	"github.com/DataDog/KubeHound/pkg/telemetry/statsd"
	"github.com/DataDog/KubeHound/pkg/telemetry/tag"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

type Dumper struct {
	directoryOutput string
	ResName         string
	collect         collector.CollectorClient
	writer          writer.DumperWriter
	ClusterName     string
}

const (
	OfflineDumpDateFormat = "2006-01-02-15-04-05"
	OfflineDumpPrefix     = "kubehound_"
)

func NewDumper() *Dumper {
	return &Dumper{}
}

func (d *Dumper) getClusterName(ctx context.Context) (string, error) {
	cluster, err := d.collect.ClusterInfo(ctx)
	if err != nil {
		return "", fmt.Errorf("collector cluster info: %w", err)
	}

	return cluster.Name, nil
}

func (d *Dumper) Initialize(ctx context.Context, collector collector.CollectorClient, compression bool, directoryOutput string) error {
	var err error

	d.directoryOutput = directoryOutput
	d.collect = collector

	d.writer, err = writer.DumperWriterFactory(ctx, compression)
	if err != nil {
		return fmt.Errorf("create collector writer: %w", err)
	}

	d.ClusterName, err = d.getClusterName(ctx)
	if err != nil {
		return err
	}

	d.ResName = path.Join(d.ClusterName, fmt.Sprintf("%s%s_%s", OfflineDumpPrefix, d.ClusterName, time.Now().Format(OfflineDumpDateFormat)))

	return d.writer.Initialize(ctx, path.Join(d.directoryOutput), d.ResName)
}

func (d *Dumper) OutputPath() string {
	return d.writer.OutputPath()
}

func (d *Dumper) IngestNode(ctx context.Context, node types.NodeType) error {
	if ok, err := preflight.CheckNode(node); !ok {

		return err
	}

	return d.processObject(ctx, node, collector.NodePath)
}

func (d *Dumper) IngestPod(ctx context.Context, pod types.PodType) error {
	if ok, err := preflight.CheckPod(pod); !ok {
		entity := tag.Entity(tag.EntityPods)
		_ = statsd.Incr(metric.CollectorSkip, append(d.collect.Tags(ctx), entity), 1)

		return err
	}

	filePath := path.Join(pod.Namespace, collector.PodPath)

	return d.processObject(ctx, pod, filePath)
}

func (d *Dumper) IngestRole(ctx context.Context, role types.RoleType) error {
	if ok, err := preflight.CheckRole(role); !ok {
		entity := tag.Entity(tag.EntityRoles)
		_ = statsd.Incr(metric.CollectorSkip, append(d.collect.Tags(ctx), entity), 1)

		return err
	}

	filePath := path.Join(role.Namespace, collector.RolesPath)

	return d.processObject(ctx, role, filePath)
}

func (d *Dumper) IngestClusterRole(ctx context.Context, clusterRole types.ClusterRoleType) error {
	if ok, err := preflight.CheckClusterRole(clusterRole); !ok {
		entity := tag.Entity(tag.EntityClusterRoles)
		_ = statsd.Incr(metric.CollectorSkip, append(d.collect.Tags(ctx), entity), 1)

		return err
	}

	return d.processObject(ctx, clusterRole, collector.ClusterRolesPath)
}

func (d *Dumper) IngestRoleBinding(ctx context.Context, roleBinding types.RoleBindingType) error {
	if ok, err := preflight.CheckRoleBinding(roleBinding); !ok {
		entity := tag.Entity(tag.EntityRolebindings)
		_ = statsd.Incr(metric.CollectorSkip, append(d.collect.Tags(ctx), entity), 1)

		return err
	}

	filePath := path.Join(roleBinding.Namespace, collector.RoleBindingsPath)

	return d.processObject(ctx, roleBinding, filePath)
}

func (d *Dumper) IngestClusterRoleBinding(ctx context.Context, clusterRoleBinding types.ClusterRoleBindingType) error {
	if ok, err := preflight.CheckClusterRoleBinding(clusterRoleBinding); !ok {
		entity := tag.Entity(tag.EntityClusterRolebindings)
		_ = statsd.Incr(metric.CollectorSkip, append(d.collect.Tags(ctx), entity), 1)

		return err
	}

	return d.processObject(ctx, clusterRoleBinding, collector.ClusterRoleBindingsPath)
}

func (d *Dumper) IngestEndpoint(ctx context.Context, endpoint types.EndpointType) error {
	if ok, err := preflight.CheckEndpoint(endpoint); !ok {
		entity := tag.Entity(tag.EntityEndpoints)
		_ = statsd.Incr(metric.CollectorSkip, append(d.collect.Tags(ctx), entity), 1)

		return err
	}

	filePath := path.Join(endpoint.Namespace, collector.EndpointPath)

	return d.processObject(ctx, endpoint, filePath)
}

// Static wrapper to dump k8s object dynamically (streams Kubernetes objects to the collector writer).
func dumpK8sObjs(ctx context.Context, operationName string, entity string, streamFunc StreamFunc) (context.Context, error) {
	log.I.Infof("Dumping %s", entity)
	span, ctx := tracer.StartSpanFromContext(ctx, operationName, tracer.Measured())
	span.SetTag(tag.EntityTag, entity)
	defer span.Finish()
	err := streamFunc(ctx)

	return ctx, err
}

func (d *Dumper) processObject(ctx context.Context, obj interface{}, filePath string) error {
	jsonData, err := json.Marshal(obj)
	if err != nil {
		return fmt.Errorf("failed to marshal Kubernetes object: %w", err)
	}

	return d.writer.Write(ctx, jsonData, filePath)
}

// Complete() is invoked by the collector when all k8s assets have been streamed.
// The function flushes all writers and waits for completion.
func (d *Dumper) Complete(ctx context.Context) error {
	d.writer.Flush(ctx)

	return nil
}

// Close() is invoked by the collector to close all handlers used to dump k8s objects.
// The function flushes all writers and close all the handlers.
func (d *Dumper) Close(ctx context.Context) error {
	d.writer.Flush(ctx)
	d.writer.Close(ctx)

	return nil
}
