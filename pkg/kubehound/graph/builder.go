package graph

import (
	"context"
	"fmt"
	"time"

	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/edge"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/converter"
	"github.com/DataDog/KubeHound/pkg/kubehound/services"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/cache"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/graphdb"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/storedb"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	"github.com/DataDog/KubeHound/pkg/telemetry/span"
	"github.com/DataDog/KubeHound/pkg/telemetry/tag"
	"github.com/DataDog/KubeHound/pkg/worker"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

// Builder handles the construction of the graph edges once vertices have been ingested via the ingestion pipeline.
type Builder struct {
	cfg     *config.KubehoundConfig
	storedb storedb.Provider
	graphdb graphdb.Provider
	cache   cache.CacheReader
	edges   *edge.Registry
}

// NewBuilder returns a new builder instance from the provided application config and service dependencies.
func NewBuilder(cfg *config.KubehoundConfig, store storedb.Provider, graph graphdb.Provider,
	cache cache.CacheReader, edges *edge.Registry) (*Builder, error) {
	n := &Builder{
		cfg:     cfg,
		storedb: store,
		graphdb: graph,
		cache:   cache,
		edges:   edges,
	}

	return n, nil
}

// HealthCheck provides a mechanism for the caller to check health of the builder dependencies.
func (b *Builder) HealthCheck(ctx context.Context) error {
	return services.HealthCheck(ctx, []services.Dependency{
		b.storedb,
		b.graphdb,
		b.cache,
	})
}

// buildEdge inserts a class of edges into the graph database.
func (b *Builder) buildEdge(ctx context.Context, label string, e edge.Builder, oic *converter.ObjectIDConverter) error {
	span, ctx := span.StartSpanFromContext(ctx, span.BuildEdge, tracer.Measured(), tracer.ResourceName(e.Label()))
	span.SetTag(tag.LabelTag, e.Label())
	var err error
	defer func() { span.Finish(tracer.WithError(err)) }()
	l := log.Logger(ctx)
	l.Info("Building edge", log.String("label", label))

	if err = e.Initialize(&b.cfg.Builder.Edge, &b.cfg.Dynamic); err != nil {
		return err
	}

	w, err := b.graphdb.EdgeWriter(ctx, e)
	if err != nil {
		return err
	}

	err = e.Stream(ctx, b.storedb, b.cache,
		func(ctx context.Context, entry types.DataContainer) error {
			insert, err := e.Processor(ctx, oic, entry)
			if err != nil {
				return err
			}

			return w.Queue(ctx, insert)
		},
		func(ctx context.Context) error {
			return w.Flush(ctx)
		})

	w.Close(ctx)

	return err
}

// buildMutating constructs all the mutating edges in the graph database.
func (b *Builder) buildMutating(ctx context.Context, oic *converter.ObjectIDConverter) error {
	for label, e := range b.edges.Mutating() {
		err := b.buildEdge(ctx, label, e, oic)
		if err != nil {
			// In case we don't want to continue and have a partial graph built, we return an error.
			// This then fails the WaitForComplete early and bubbles up to main.
			if b.cfg.Builder.StopOnError {
				return fmt.Errorf("building mutating edge %s: %w", label, err)
			}
			// Otherwise, by returning nil AND printing an error, we explicitly tell the user that something is going wrong.
			// Since the issue might not be easy or even possible for the user to fix, we still want to be able to provide _some_
			// values to the user (permissions of the users etc...)
			// TODO(#ASENG-512): Add an error handling framework to accumulate all errors and display them to the user in an user friendly way
			// l.Errorf("Failed to create a mutating edge (type: %s). The created graph will be INCOMPLETE (change `builder.stop_on_error` to abort or error instead)", e.Name())

			return nil
		}
	}

	return nil
}

// buildSimple constructs all the simple edges in the graph database.
func (b *Builder) buildSimple(ctx context.Context, oic *converter.ObjectIDConverter) error {
	l := log.Logger(ctx)
	l.Info("Creating edge builder worker pool")
	wp, err := worker.PoolFactory(b.cfg.Builder.Edge.WorkerPoolSize, b.cfg.Builder.Edge.WorkerPoolCapacity)
	if err != nil {
		return fmt.Errorf("graph builder worker pool create: %w", err)
	}

	workCtx, err := wp.Start(ctx)
	if err != nil {
		return fmt.Errorf("graph builder worker pool start: %w", err)
	}

	for label, e := range b.edges.Simple() {
		e := e
		label := label

		wp.Submit(func() error {
			err := b.buildEdge(workCtx, label, e, oic)
			if err != nil {
				// l.Errorf("building simple edge %s: %v", label, err)
				// In case we don't want to continue and have a partial graph built, we return an error.
				// This then fails the WaitForComplete early and bubbles up to main.
				if b.cfg.Builder.StopOnError {
					return err
				}
				// Otherwise, by returning nil AND printing an error, we explicitly tell the user that something is going wrong.
				// Since the issue might not be easy or even possible for the user to fix, we still want to be able to provide _some_
				// values to the user (permissions of the users etc...)
				// TODO(#ASENG-512): Add an error handling framework to accumulate all errors and display them to the user in an user friendly way
				// l.Errorf("Failed to create a simple edge (type: %s). The created graph will be INCOMPLETE (change `builder.stop_on_error` to abort or error instead)", e.Name())

				return nil
			}

			return nil
		})
	}

	err = wp.WaitForComplete()
	if err != nil {
		return err
	}

	return nil
}

// buildDependent constructs all the dependent edges in the graph database.
func (b *Builder) buildDependent(ctx context.Context, oic *converter.ObjectIDConverter) error {
	for label, e := range b.edges.Dependent() {
		err := b.buildEdge(ctx, label, e, oic)
		if err != nil {
			// In case we don't want to continue and have a partial graph built, we return an error.
			// This then fails the WaitForComplete early and bubbles up to main.
			if b.cfg.Builder.StopOnError {
				return fmt.Errorf("building dependent edge %s: %w", label, err)
			}
			// Otherwise, by returning nil AND printing an error, we explicitly tell the user that something is going wrong.
			// Since the issue might not be easy or even possible for the user to fix, we still want to be able to provide _some_
			// values to the user (permissions of the users etc...)
			// TODO(#ASENG-512): Add an error handling framework to accumulate all errors and display them to the user in an user friendly way
			// l.Errorf("Failed to create a dependent edge (type: %s). The created graph will be INCOMPLETE (change `builder.stop_on_error` to abort or error instead)", e.Name())

			return nil
		}
	}

	return nil
}

// Run constructs all the registered edges in the graph database.
// NOTE: edges are constructed in parallel using a worker pool with properties configured via the top-level KubeHound config.
func (b *Builder) Run(ctx context.Context) error {
	l := log.Trace(ctx)
	oic := converter.NewObjectID(b.cache)

	if b.cfg.Builder.Edge.LargeClusterOptimizations {
		log.Trace(ctx).Warnf("Using large cluster optimizations in graph construction")
	}

	// Mutating edges must be built first, sequentially
	l.Info("Starting mutating edge construction")
	if err := b.buildMutating(ctx, oic); err != nil {
		return err
	}

	// Simple edges can be built in parallel
	l.Info("Starting simple edge construction")
	if err := b.buildSimple(ctx, oic); err != nil {
		return err
	}

	// Dependent edges must be built last, sequentially
	l.Info("Starting dependent edge construction")
	if err := b.buildDependent(ctx, oic); err != nil {
		return err
	}

	l.Info("Completed edge construction")

	return nil
}

// buildGraph will construct the attack graph by calculating and inserting all registered edges in parallel.
// All I/O operations are performed asynchronously.
func BuildGraph(outer context.Context, cfg *config.KubehoundConfig, storedb storedb.Provider,
	graphdb graphdb.Provider, cache cache.CacheReader) error {
	l := log.Logger(outer)
	start := time.Now()
	span, ctx := span.SpanRunFromContext(outer, span.BuildGraph)
	var err error
	defer func() { span.Finish(tracer.WithError(err)) }()

	l.Info("Loading graph edge definitions")
	edges := edge.Registered()
	if err = edges.Verify(); err != nil {
		return fmt.Errorf("edge registry verification: %w", err)
	}

	l.Info("Loading graph builder")
	builder, err := NewBuilder(cfg, storedb, graphdb, cache, edges)
	if err != nil {
		return fmt.Errorf("graph builder creation: %w", err)
	}

	l.Info("Running dependency health checks")
	if err := builder.HealthCheck(ctx); err != nil {
		return fmt.Errorf("graph builder dependency health check: %w", err)
	}

	l.Info("Constructing graph")
	if err := builder.Run(ctx); err != nil {
		return fmt.Errorf("graph builder edge calculation: %w", err)
	}

	l.Info("Completed graph construction", log.Duration("duration", time.Since(start)))

	return nil
}
