package collector

import (
	"context"
	"encoding/json"
	"fmt"
	"path"

	"time"

	"github.com/DataDog/KubeHound/pkg/collector/writer"
	"github.com/DataDog/KubeHound/pkg/globals/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/ingestor/preflight"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	"github.com/DataDog/KubeHound/pkg/telemetry/span"
	"github.com/DataDog/KubeHound/pkg/telemetry/tag"
	"github.com/DataDog/KubeHound/pkg/worker"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

type OfflineDump struct {
	directoryOutput string
	ResName         string
	collect         CollectorClient
	writer          writer.CollectorWriter
	ClusterName     string
}

const (
	OfflineDumpDateFormat = "2006-01-02-15-04-05"
	OfflineDumpPrefix     = "kubehound_"
)

func NewOfflineDump() *OfflineDump {
	return &OfflineDump{}
}

func (o *OfflineDump) getClusterName(ctx context.Context) (string, error) {
	cluster, err := o.collect.ClusterInfo(ctx)
	if err != nil {
		return "", fmt.Errorf("collector cluster info: %w", err)
	}
	return cluster.Name, nil
}

func (o *OfflineDump) Initialize(ctx context.Context, collector CollectorClient, compression bool, directoryOutput string) error {
	var err error

	o.directoryOutput = directoryOutput
	o.collect = collector

	o.writer = &writer.FileWriter{}
	// if compression is enabled, create the tar.gz file
	if compression {
		log.I.Infof("Compression enabled")
		o.writer = &writer.TarWriter{}
	}
	o.ClusterName, err = o.getClusterName(ctx)
	if err != nil {
		return err
	}

	o.ResName = path.Join(o.ClusterName, fmt.Sprintf("%s%s_%s", OfflineDumpPrefix, o.ClusterName, time.Now().Format(OfflineDumpDateFormat)))

	return o.writer.Initialize(ctx, path.Join(o.directoryOutput), o.ResName)
}

func (o *OfflineDump) GetOutputPath() string {
	return o.writer.GetOutputPath()
}

func (o *OfflineDump) IngestNode(ctx context.Context, node types.NodeType) error {
	if ok, err := preflight.CheckNode(node); !ok {
		return err
	}

	return o.processObject(ctx, node, nodePath)
}

func (i *OfflineDump) dumpNodes(ctx context.Context) error {
	log.I.Info("Dumping nodes")
	span, ctx := tracer.StartSpanFromContext(ctx, span.CollectorDump, tracer.Measured())
	span.SetTag(tag.EntityTag, tag.EntityNodes)
	defer span.Finish()

	return i.collect.StreamNodes(ctx, i)
}

func (o *OfflineDump) IngestPod(ctx context.Context, pod types.PodType) error {
	if ok, err := preflight.CheckPod(pod); !ok {
		return err
	}

	filePath := path.Join(pod.Namespace, podPath)

	return o.processObject(ctx, pod, filePath)
}

func (i *OfflineDump) dumpPods(ctx context.Context) error {
	log.I.Info("Dumping pods")
	return i.collect.StreamPods(ctx, i)
}

func (o *OfflineDump) IngestRole(ctx context.Context, role types.RoleType) error {
	if ok, err := preflight.CheckRole(role); !ok {
		return err
	}

	filePath := path.Join(role.Namespace, rolesPath)

	return o.processObject(ctx, role, filePath)
}

func (i *OfflineDump) dumpRoles(ctx context.Context) error {
	log.I.Info("Dumping roles")
	return i.collect.StreamRoles(ctx, i)
}

func (o *OfflineDump) IngestClusterRole(ctx context.Context, clusterRole types.ClusterRoleType) error {
	if ok, err := preflight.CheckClusterRole(clusterRole); !ok {
		return err
	}

	return o.processObject(ctx, clusterRole, clusterRolesPath)
}

func (i *OfflineDump) dumpClusterRoles(ctx context.Context) error {
	log.I.Info("Dumping cluster roles")
	return i.collect.StreamClusterRoles(ctx, i)
}

func (o *OfflineDump) IngestRoleBinding(ctx context.Context, roleBinding types.RoleBindingType) error {
	if ok, err := preflight.CheckRoleBinding(roleBinding); !ok {
		return err
	}

	filePath := path.Join(roleBinding.Namespace, roleBindingsPath)

	return o.processObject(ctx, roleBinding, filePath)
}

func (i *OfflineDump) DumpRoleBindings(ctx context.Context) error {
	log.I.Info("Dumping role bindings")
	return i.collect.StreamRoleBindings(ctx, i)
}

func (o *OfflineDump) IngestClusterRoleBinding(ctx context.Context, clusterRoleBinding types.ClusterRoleBindingType) error {
	if ok, err := preflight.CheckClusterRoleBinding(clusterRoleBinding); !ok {
		return err
	}

	return o.processObject(ctx, clusterRoleBinding, clusterRoleBindingsPath)
}

func (i *OfflineDump) dumpClusterRoleBinding(ctx context.Context) error {
	log.I.Info("Dumping cluster role bindings")
	return i.collect.StreamClusterRoleBindings(ctx, i)
}

func (o *OfflineDump) IngestEndpoint(ctx context.Context, endpoint types.EndpointType) error {
	if ok, err := preflight.CheckEndpoint(endpoint); !ok {
		return err
	}

	filePath := path.Join(endpoint.Namespace, endpointPath)

	return o.processObject(ctx, endpoint, filePath)
}

func (i *OfflineDump) dumpEnpoints(ctx context.Context) error {
	log.I.Info("Dumping endpoints")
	return i.collect.StreamEndpoints(ctx, i)
}

func (i *OfflineDump) DumpK8sObjects(ctx context.Context, workerNumber int) error {
	span, ctx := tracer.StartSpanFromContext(ctx, span.CollectorTarWriterClose, tracer.Measured())
	span.SetTag(tag.CollectorWorkerNumber, workerNumber)
	defer span.Finish()

	wp, err := worker.PoolFactory(workerNumber, 1)
	if err != nil {
		return fmt.Errorf("create worker pool: %w", err)
	}

	_, err = wp.Start(ctx)
	if err != nil {
		return fmt.Errorf("group worker pool start: %w", err)
	}

	defer i.writer.Close(ctx)

	wp.Submit(func() error {
		return i.dumpNodes(ctx)
	})

	wp.Submit(func() error {
		return i.dumpPods(ctx)
	})

	wp.Submit(func() error {
		return i.dumpRoles(ctx)
	})

	wp.Submit(func() error {
		return i.dumpClusterRoles(ctx)
	})

	wp.Submit(func() error {
		return i.DumpRoleBindings(ctx)
	})

	wp.Submit(func() error {
		return i.dumpClusterRoleBinding(ctx)
	})

	wp.Submit(func() error {
		return i.dumpEnpoints(ctx)
	})

	return wp.WaitForComplete()
}

func (o *OfflineDump) processObject(ctx context.Context, obj interface{}, filePath string) error {
	jsonData, err := json.Marshal(obj)
	if err != nil {
		return fmt.Errorf("failed to marshal Kubernetes object: %w", err)
	}

	o.writer.Write(ctx, jsonData, filePath)

	return err
}

// completeCallback is invoked by the collector when all pods have been streamed.
// The function flushes all writers and waits for completion.
func (o *OfflineDump) Complete(ctx context.Context) error {
	o.writer.Flush(ctx)
	return nil
}

// completeCallback is invoked by the collector when all pods have been streamed.
// The function flushes all writers and close all the handlers.
func (o *OfflineDump) Close(ctx context.Context) error {
	o.writer.Flush(ctx)
	o.writer.Close(ctx)
	return nil
}
