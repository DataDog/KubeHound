package dump

import (
	"context"
	"encoding/json"
	"fmt"
	"path"

	"time"

	"github.com/DataDog/KubeHound/pkg/collector"
	"github.com/DataDog/KubeHound/pkg/dump/writer"
	"github.com/DataDog/KubeHound/pkg/globals/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/ingestor/preflight"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	"github.com/DataDog/KubeHound/pkg/telemetry/metric"
	"github.com/DataDog/KubeHound/pkg/telemetry/span"
	"github.com/DataDog/KubeHound/pkg/telemetry/statsd"
	"github.com/DataDog/KubeHound/pkg/telemetry/tag"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

var _ collector.GenericIngestor = (*DumpIngestor)(nil)

type DumpIngestor struct {
	directoryOutput string
	ResName         string
	collector       collector.CollectorClient
	writer          writer.DumperWriter
	ClusterName     string
}

const (
	OfflineDumpDateFormat = "2006-01-02-15-04-05"
	OfflineDumpPrefix     = "kubehound_"
)

// ./<clusterName>/kubehound_<clusterName>_<date>
func dumpIngestorResName(clusterName string) string {
	return path.Join(clusterName, fmt.Sprintf("%s%s_%s", OfflineDumpPrefix, clusterName, time.Now().Format(OfflineDumpDateFormat)))
}

func NewDumpIngestor(ctx context.Context, collector collector.CollectorClient, compression bool, directoryOutput string) (*DumpIngestor, error) {
	// Generate path for the dump
	clusterName, err := getClusterName(ctx, collector)
	if err != nil {
		return nil, err
	}

	resName := dumpIngestorResName(clusterName)

	dumpWriter, err := writer.DumperWriterFactory(ctx, compression, directoryOutput, resName)
	if err != nil {
		return nil, fmt.Errorf("create collector writer: %w", err)
	}

	return &DumpIngestor{
		directoryOutput: directoryOutput,
		collector:       collector,
		ClusterName:     clusterName,
		writer:          dumpWriter,
		ResName:         resName,
	}, nil
}

func getClusterName(ctx context.Context, collector collector.CollectorClient) (string, error) {
	cluster, err := collector.ClusterInfo(ctx)
	if err != nil {
		return "", fmt.Errorf("collector cluster info: %w", err)
	}

	return cluster.Name, nil
}

func (d *DumpIngestor) OutputPath() string {
	return d.writer.OutputPath()
}

func (d *DumpIngestor) IngestNode(ctx context.Context, node types.NodeType) error {
	if ok, err := preflight.CheckNode(node); !ok {

		return err
	}

	return d.processObject(ctx, node, collector.NodePath)
}

func ingestPodPath(pod types.PodType) string {
	return path.Join(pod.Namespace, collector.PodPath)
}

func (d *DumpIngestor) IngestPod(ctx context.Context, pod types.PodType) error {
	if ok, err := preflight.CheckPod(pod); !ok {
		entity := tag.Entity(tag.EntityPods)
		_ = statsd.Incr(metric.CollectorSkip, append(d.collector.Tags(ctx), entity), 1)

		return err
	}

	filePath := ingestPodPath(pod)

	return d.processObject(ctx, pod, filePath)
}

func ingestRolePath(roleBinding types.RoleType) string {
	return path.Join(roleBinding.Namespace, collector.RolesPath)
}

func (d *DumpIngestor) IngestRole(ctx context.Context, role types.RoleType) error {
	if ok, err := preflight.CheckRole(role); !ok {
		entity := tag.Entity(tag.EntityRoles)
		_ = statsd.Incr(metric.CollectorSkip, append(d.collector.Tags(ctx), entity), 1)

		return err
	}

	filePath := ingestRolePath(role)

	return d.processObject(ctx, role, filePath)
}

func (d *DumpIngestor) IngestClusterRole(ctx context.Context, clusterRole types.ClusterRoleType) error {
	if ok, err := preflight.CheckClusterRole(clusterRole); !ok {
		entity := tag.Entity(tag.EntityClusterRoles)
		_ = statsd.Incr(metric.CollectorSkip, append(d.collector.Tags(ctx), entity), 1)

		return err
	}

	return d.processObject(ctx, clusterRole, collector.ClusterRolesPath)
}

func ingestRoleBindingPath(roleBinding types.RoleBindingType) string {
	return path.Join(roleBinding.Namespace, collector.RoleBindingsPath)
}

func (d *DumpIngestor) IngestRoleBinding(ctx context.Context, roleBinding types.RoleBindingType) error {
	if ok, err := preflight.CheckRoleBinding(roleBinding); !ok {
		entity := tag.Entity(tag.EntityRolebindings)
		_ = statsd.Incr(metric.CollectorSkip, append(d.collector.Tags(ctx), entity), 1)

		return err
	}

	filePath := ingestRoleBindingPath(roleBinding)

	return d.processObject(ctx, roleBinding, filePath)
}

func (d *DumpIngestor) IngestClusterRoleBinding(ctx context.Context, clusterRoleBinding types.ClusterRoleBindingType) error {
	if ok, err := preflight.CheckClusterRoleBinding(clusterRoleBinding); !ok {
		entity := tag.Entity(tag.EntityClusterRolebindings)
		_ = statsd.Incr(metric.CollectorSkip, append(d.collector.Tags(ctx), entity), 1)

		return err
	}

	return d.processObject(ctx, clusterRoleBinding, collector.ClusterRoleBindingsPath)
}

func ingestEndpointPath(endpoint types.EndpointType) string {
	return path.Join(endpoint.Namespace, collector.EndpointPath)
}

func (d *DumpIngestor) IngestEndpoint(ctx context.Context, endpoint types.EndpointType) error {
	if ok, err := preflight.CheckEndpoint(endpoint); !ok {
		entity := tag.Entity(tag.EntityEndpoints)
		_ = statsd.Incr(metric.CollectorSkip, append(d.collector.Tags(ctx), entity), 1)

		return err
	}

	filePath := ingestEndpointPath(endpoint)

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

func (d *DumpIngestor) DumpK8sObjects(ctx context.Context) error {
	spanDump, ctx := tracer.StartSpanFromContext(ctx, span.CollectorDump, tracer.Measured())
	defer spanDump.Finish()

	ctx, pipeline, err := newPipelineDumpIngestor(ctx, d)
	if err != nil {
		return fmt.Errorf("create pipeline ingestor: %w", err)
	}
	spanDump.SetTag(tag.DumperWorkerNumberTag, pipeline.WorkerNumber)

	err = pipeline.Run(ctx)
	if err != nil {
		return fmt.Errorf("run pipeline ingestor: %w", err)
	}

	return pipeline.Wait(ctx)
}

func marshalK8sObj(obj interface{}) ([]byte, error) {
	jsonData, err := json.Marshal(obj)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal Kubernetes object: %w", err)
	}

	return jsonData, nil
}

func (d *DumpIngestor) processObject(ctx context.Context, obj interface{}, filePath string) error {
	jsonData, err := marshalK8sObj(obj)
	if err != nil {
		return err
	}

	return d.writer.Write(ctx, jsonData, filePath)
}

// Complete() is invoked by the collector when all k8s assets have been streamed.
// The function flushes all writers and waits for completion.
func (d *DumpIngestor) Complete(ctx context.Context) error {
	d.writer.Flush(ctx)

	return nil
}

// Close() is invoked by the collector to close all handlers used to dump k8s objects.
// The function flushes all writers and close all the handlers.
func (d *DumpIngestor) Close(ctx context.Context) error {
	d.writer.Flush(ctx)
	d.writer.Close(ctx)

	return nil
}
