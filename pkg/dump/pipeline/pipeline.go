package pipeline

import (
	"context"
	"errors"
	"fmt"

	"github.com/DataDog/KubeHound/pkg/collector"
	"github.com/DataDog/KubeHound/pkg/dump/writer"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	"github.com/DataDog/KubeHound/pkg/telemetry/span"
	"github.com/DataDog/KubeHound/pkg/telemetry/tag"
	"github.com/DataDog/KubeHound/pkg/worker"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

type StreamFunc func(context.Context) error

type DumpIngestorPipeline struct {
	operationName string
	entity        string
	streamFunc    StreamFunc
}

// dumpIngestorSequence returns the pipeline sequence for dumping k8s object (can be multi-threaded depending on the writer used)
func dumpIngestorSequence(collector collector.CollectorClient, writer writer.DumperWriter) []DumpIngestorPipeline {
	return []DumpIngestorPipeline{
		{
			operationName: span.DumperNodes,
			entity:        tag.EntityNodes,
			streamFunc: func(ctx context.Context) error {
				return collector.StreamNodes(ctx, NewNodeIngestor(ctx, writer))
			},
		},
		{
			operationName: span.DumperPods,
			entity:        tag.EntityPods,
			streamFunc: func(ctx context.Context) error {
				return collector.StreamPods(ctx, NewPodIngestor(ctx, writer))
			},
		},
		{
			operationName: span.DumperRoles,
			entity:        tag.EntityRoles,
			streamFunc: func(ctx context.Context) error {
				return collector.StreamRoles(ctx, NewRoleIngestor(ctx, writer))
			},
		},
		{
			operationName: span.DumperClusterRoles,
			entity:        tag.EntityClusterRoles,
			streamFunc: func(ctx context.Context) error {
				return collector.StreamClusterRoles(ctx, NewClusterRoleIngestor(ctx, writer))
			},
		},
		{
			operationName: span.DumperRoleBindings,
			entity:        tag.EntityRolebindings,
			streamFunc: func(ctx context.Context) error {
				return collector.StreamRoleBindings(ctx, NewRoleBindingIngestor(ctx, writer))
			},
		},
		{
			operationName: span.DumperClusterRoleBindings,
			entity:        tag.EntityClusterRolebindings,
			streamFunc: func(ctx context.Context) error {
				return collector.StreamClusterRoleBindings(ctx, NewClusterRoleBindingIngestor(ctx, writer))
			},
		},
		{
			operationName: span.DumperEndpoints,
			entity:        tag.EntityEndpoints,
			streamFunc: func(ctx context.Context) error {
				return collector.StreamEndpoints(ctx, NewEndpointIngestor(ctx, writer))
			},
		},
	}
}

// dumpIngestorClosingSequence returns the pipeline sequence for closing the dumping sequence (this pipeline is single-threaded)
func dumpIngestorClosingSequence(collector collector.CollectorClient, writer writer.DumperWriter) []DumpIngestorPipeline {
	return []DumpIngestorPipeline{
		{
			operationName: span.DumperMetadata,
			entity:        "Metadata",
			streamFunc: func(ctx context.Context) error {
				return collector.ComputeMetadata(ctx, NewMetadataIngestor(ctx, writer))
			},
		},
	}
}

// PipelineDumpIngestor is a parallelized pipeline based ingestor implementation.
type PipelineDumpIngestor struct {
	sequence        []DumpIngestorPipeline
	closingSequence []DumpIngestorPipeline
	wp              worker.WorkerPool
	WorkerNumber    int
}

func NewPipelineDumpIngestor(ctx context.Context, collector collector.CollectorClient, writer writer.DumperWriter) (context.Context, *PipelineDumpIngestor, error) {
	sequence := dumpIngestorSequence(collector, writer)
	cleanupSequence := dumpIngestorClosingSequence(collector, writer)

	// Getting the number of workers from the writer to setup multi-threading if possible
	workerNumber := writer.WorkerNumber()
	// Set the number of workers to the number of differents entities (roles, pods, ...)
	if workerNumber == 0 {
		workerNumber = len(sequence)
	}

	if workerNumber > 1 {
		log.I.Infof("Multi-threading enabled: %d workers", workerNumber)
	}

	// Setting up the worker pool with multi-threading if possible
	// enable for raw file writer
	// disable for targz writer (not thread safe)
	bufferCapacity := 1
	wp, err := worker.PoolFactory(workerNumber, bufferCapacity)
	if err != nil {
		return nil, nil, fmt.Errorf("create worker pool: %w", err)
	}

	ctx, err = wp.Start(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("group worker pool start: %w", err)
	}

	return ctx, &PipelineDumpIngestor{
		wp:              wp,
		sequence:        sequence,
		closingSequence: cleanupSequence,
		WorkerNumber:    workerNumber,
	}, nil

}

func (p *PipelineDumpIngestor) Run(ctx context.Context) error {
	var err error
	for _, v := range p.sequence {
		v := v
		p.wp.Submit(func() error {
			errDump := dumpK8sObjs(ctx, v.operationName, v.entity, v.streamFunc)
			if errDump != nil {
				err = errors.Join(err, errDump)
			}

			return err
		})
	}

	return err
}

func (p *PipelineDumpIngestor) WaitAndClose(ctx context.Context) error {
	err := p.wp.WaitForComplete()
	if err != nil {
		return fmt.Errorf("wait for complete: %w", err)
	}

	for _, v := range p.closingSequence {
		v := v
		errDump := dumpK8sObjs(ctx, v.operationName, v.entity, v.streamFunc)
		if errDump != nil {
			err = errors.Join(err, errDump)
		}
	}

	return err
}

// Static wrapper to dump k8s object dynamically (streams Kubernetes objects to the collector writer).
func dumpK8sObjs(ctx context.Context, operationName string, entity string, streamFunc StreamFunc) error {
	log.I.Infof("Dumping %s", entity)
	span, ctx := tracer.StartSpanFromContext(ctx, operationName, tracer.Measured())
	span.SetTag(tag.EntityTag, entity)
	var err error
	defer func() { span.Finish(tracer.WithError(err)) }()
	err = streamFunc(ctx)
	log.I.Infof("Dumping %s done", entity)

	return err
}
