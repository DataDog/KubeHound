package dumper

import (
	"context"
	"errors"
	"fmt"

	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	"github.com/DataDog/KubeHound/pkg/telemetry/span"
	"github.com/DataDog/KubeHound/pkg/telemetry/tag"
	"github.com/DataDog/KubeHound/pkg/worker"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

type StreamFunc func(context.Context) error

type DumperPipeline struct {
	operationName string
	entity        string
	streamFunc    StreamFunc
}

// dumperSequence returns the pipeline sequence for dumping k8s object (can be multi-threaded depending on the writer used)
func dumperSequence(d *Dumper) []DumperPipeline {
	return []DumperPipeline{
		{
			operationName: span.DumperNodes,
			entity:        tag.EntityNodes,
			streamFunc: func(ctx context.Context) error {
				return d.collect.StreamNodes(ctx, d)
			},
		},
		{
			operationName: span.DumperPods,
			entity:        tag.EntityPods,
			streamFunc: func(ctx context.Context) error {
				return d.collect.StreamPods(ctx, d)
			},
		},
		{
			operationName: span.DumperRoles,
			entity:        tag.EntityRoles,
			streamFunc: func(ctx context.Context) error {
				return d.collect.StreamRoles(ctx, d)
			},
		},
		{
			operationName: span.DumperClusterRoles,
			entity:        tag.EntityClusterRoles,
			streamFunc: func(ctx context.Context) error {
				return d.collect.StreamClusterRoles(ctx, d)
			},
		},
		{
			operationName: span.DumperRoleBindings,
			entity:        tag.EntityRolebindings,
			streamFunc: func(ctx context.Context) error {
				return d.collect.StreamRoleBindings(ctx, d)
			},
		},
		{
			operationName: span.DumperClusterRoleBindings,
			entity:        tag.EntityClusterRolebindings,
			streamFunc: func(ctx context.Context) error {
				return d.collect.StreamClusterRoleBindings(ctx, d)
			},
		},
		{
			operationName: span.DumperEndpoints,
			entity:        tag.EntityEndpoints,
			streamFunc: func(ctx context.Context) error {
				return d.collect.StreamEndpoints(ctx, d)
			},
		},
	}
}

// PipelineDumper is a parallelized pipeline based ingestor implementation.
type PipelineDumper struct {
	sequence     []DumperPipeline
	wp           worker.WorkerPool
	WorkerNumber int
}

func newPipelineDumper(ctx context.Context, d *Dumper) (context.Context, *PipelineDumper, error) {
	sequence := dumperSequence(d)

	// Getting the number of workers from the writer to setup multi-threading if possible
	workerNumber := d.writer.WorkerNumber()
	if workerNumber == 0 {
		log.I.Infof("Multi-threading enabled: %d workers", workerNumber)
		workerNumber = len(sequence)
	}

	// Setting up the worker pool with multi-threading if possible
	// enable for raw file writer
	// disable for targz writer (not thread safe)
	wp, err := worker.PoolFactory(workerNumber, 1)
	if err != nil {
		return nil, nil, fmt.Errorf("create worker pool: %w", err)
	}

	ctx, err = wp.Start(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("group worker pool start: %w", err)
	}

	return ctx, &PipelineDumper{
		wp:           wp,
		sequence:     sequence,
		WorkerNumber: workerNumber,
	}, nil

}

func (p *PipelineDumper) Run(ctx context.Context) error {
	var err error
	for _, v := range p.sequence {
		v := v
		p.wp.Submit(func() error {
			_, errDump := dumpK8sObjs(ctx, v.operationName, v.entity, v.streamFunc)
			if errDump != nil {
				err = errors.Join(err, errDump)
			}

			return err
		})
	}

	return err
}

func (p *PipelineDumper) Wait(ctx context.Context) error {
	return p.wp.WaitForComplete()
}

func (d *Dumper) DumpK8sObjects(ctx context.Context) error {
	spanDump, ctx := tracer.StartSpanFromContext(ctx, span.CollectorDump, tracer.Measured())
	defer spanDump.Finish()

	ctx, pipeline, err := newPipelineDumper(ctx, d)
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