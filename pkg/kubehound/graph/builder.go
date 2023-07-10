package graph

import (
	"context"
	"fmt"

	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/DataDog/KubeHound/pkg/globals"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/edge"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/converter"
	"github.com/DataDog/KubeHound/pkg/kubehound/services"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/cache"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/graphdb"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/storedb"
	"github.com/DataDog/KubeHound/pkg/telemetry"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	"github.com/DataDog/KubeHound/pkg/worker"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

// Builder handles the construction of the graph edges once vertices have been ingested via the ingestion pipeline.
type Builder struct {
	cfg     *config.KubehoundConfig
	storedb storedb.Provider
	graphdb graphdb.Provider
	cache   cache.CacheReader
	edges   edge.Registry
}

// NewBuilder returns a new builder instance from the provided application config and service dependencies.
func NewBuilder(cfg *config.KubehoundConfig, store storedb.Provider, graph graphdb.Provider,
	cache cache.CacheReader, edges edge.Registry) (*Builder, error) {

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
// NOTE: function is blocking and expected to be called from within a goroutine.
func (b *Builder) buildEdge(ctx context.Context, e edge.Builder, oic *converter.ObjectIDConverter) error {
	tags := append(telemetry.BaseTags, telemetry.TagTypeJanusGraph)
	w, err := b.graphdb.EdgeWriter(ctx, e, graphdb.WithTags(tags))
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

// Run constructs all the registered edges in the graph database.
// NOTE: edges are constructed in parallel using a worker pool with properties configured via the top-level KubeHound config.
func (b *Builder) Run(ctx context.Context) error {
	span, ctx := tracer.StartSpanFromContext(ctx, telemetry.SpanOperationRun, tracer.Measured())
	defer span.Finish()
	l := log.Trace(ctx, log.WithComponent(globals.BuilderComponent))

	// Edges can be built in parallel
	l.Info("Creating edge builder worker pool")
	wp, err := worker.PoolFactory(b.cfg.Builder.Edge.WorkerPoolSize, b.cfg.Builder.Edge.WorkerPoolCapacity)
	if err != nil {
		return fmt.Errorf("graph builder worker pool create: %w", err)
	}

	workCtx, err := wp.Start(ctx)
	if err != nil {
		return fmt.Errorf("graph builder worker pool start: %w", err)
	}

	l.Info("Starting edge construction")
	if b.cfg.Builder.Edge.LargeClusterOptimizations {
		log.I.Warnf("Using large cluster optimizations in graph construction")
	}

	oic := converter.NewObjectID(b.cache)
	for label, e := range b.edges {
		e := e
		label := label

		wp.Submit(func() error {
			l.Infof("Building edge %s", label)

			if err := e.Initialize(&b.cfg.Builder.Edge); err != nil {
				l.Errorf("initializing edge %s: %v", label, err)
				return err
			}

			err := b.buildEdge(workCtx, e, oic)
			if err != nil {
				l.Errorf("building edge %s: %v", label, err)
				return err
			}

			return nil
		})
	}

	err = wp.WaitForComplete()
	if err != nil {
		return err
	}

	l.Info("Completed edge construction")
	return nil
}
