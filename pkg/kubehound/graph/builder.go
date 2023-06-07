package graph

import (
	"context"
	"fmt"

	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/DataDog/KubeHound/pkg/globals"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/edge"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/vertex"
	"github.com/DataDog/KubeHound/pkg/kubehound/services"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/cache"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/graphdb"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/storedb"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	"github.com/DataDog/KubeHound/pkg/worker"
)

// Builder handles the construction of the graph edges once vertices have been ingested via the ingestion pipeline.
type Builder struct {
	cfg      *config.KubehoundConfig
	storedb  storedb.Provider
	graphdb  graphdb.Provider
	cache    cache.CacheReader
	edges    edge.EdgeRegistry
	vertices vertex.VertexRegistry
}

// NewBuilder returns a new builder instance from the provided application config and service dependencies.
func NewBuilder(cfg *config.KubehoundConfig, store storedb.Provider, graph graphdb.Provider,
	cache cache.CacheReader, edges edge.EdgeRegistry, vertices vertex.VertexRegistry) (*Builder, error) {

	n := &Builder{
		cfg:      cfg,
		storedb:  store,
		graphdb:  graph,
		cache:    cache,
		edges:    edges,
		vertices: vertices,
	}

	return n, nil
}

// HealthCheck provides a mechanism for the caller to check health of the builder dependencies.
func (b *Builder) HealthCheck(ctx context.Context) error {
	return services.HealthCheck(ctx, []services.Dependency{
		b.storedb,
		b.graphdb,
	})
}

// buildVertex inserts a class of vertices into the graph database.
func (b *Builder) buildVertex(ctx context.Context, v vertex.QueryBuilder) error {
	w, err := b.graphdb.VertexWriter(ctx, v)
	if err != nil {
		return err
	}

	err = v.Stream(ctx, b.storedb, b.cache,
		func(ctx context.Context, entry types.DataContainer) error {
			processed, err := v.Processor(ctx, entry)
			// TODO option for skip write if signalled by processor

			if err != nil {
				// TODO tolerate errors
				return err
			}

			return w.Queue(ctx, processed)

		},
		func(ctx context.Context) error {
			return w.Flush(ctx)
		})

	w.Close(ctx)

	return err
}

// buildEdge inserts a class of edges into the graph database.
// NOTE: function is blocking and expected to be called from within a goroutine.
func (b *Builder) buildEdge(ctx context.Context, e edge.Builder) error {
	w, err := b.graphdb.EdgeWriter(ctx, e)
	if err != nil {
		return err
	}

	err = e.Stream(ctx, b.storedb, b.cache,
		func(ctx context.Context, entry types.DataContainer) error {
			processed, err := e.Processor(ctx, entry)
			// TODO option for skip write if signalled by processor

			if err != nil {
				// TODO tolerate errors
				return err
			}

			return w.Queue(ctx, processed)

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
	l := log.Trace(ctx, log.WithComponent(globals.BuilderComponent))

	// Vertices can have dependencies so must be built in sequence
	l.Info("Starting vertex construction")
	for label, v := range b.vertices {
		l.Infof("Building vertex %s", label)

		err := b.buildVertex(ctx, v)
		if err != nil {
			l.Errorf("building verrtex %s: %v", label, err)
			return err
		}

		return nil
	}

	// Edges can be built in parallel
	l.Info("Creating edge builder worker pool")
	wp, err := worker.PoolFactory(b.cfg)
	if err != nil {
		return fmt.Errorf("graph builder worker pool create: %w", err)
	}

	workCtx, err := wp.Start(ctx)
	if err != nil {
		return fmt.Errorf("graph builder worker pool start: %w", err)
	}

	l.Info("Starting edge construction")

	for label, e := range b.edges {
		e := e
		label := label

		wp.Submit(func() error {
			l.Infof("Building edge %s", label)

			err := b.buildEdge(workCtx, e)
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
