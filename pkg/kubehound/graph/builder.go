package graph

import (
	"context"
	"fmt"
	"time"

	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/DataDog/KubeHound/pkg/globals"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/edge"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/path"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/services"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/cache"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/graphdb"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/storedb"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	"github.com/DataDog/KubeHound/pkg/worker"
)

// Builder handles the construction of the graph edges once vertices have been ingested via the ingestion pipeline.
type Builder struct {
	cfg     *config.KubehoundConfig
	storedb storedb.Provider
	graphdb graphdb.Provider
	cache   cache.CacheReader
	edges   edge.Registry
	paths   path.Registry
}

// NewBuilder returns a new builder instance from the provided application config and service dependencies.
func NewBuilder(cfg *config.KubehoundConfig, store storedb.Provider, graph graphdb.Provider,
	cache cache.CacheReader, edges edge.Registry, paths path.Registry) (*Builder, error) {

	n := &Builder{
		cfg:     cfg,
		storedb: store,
		graphdb: graph,
		cache:   cache,
		edges:   edges,
		paths:   paths,
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

// buildPath inserts a class of paths (combination of new vertices and edges) into the graph database.
func (b *Builder) buildPath(ctx context.Context, p path.Builder) error {
	w, err := b.graphdb.PathWriter(ctx, p)
	if err != nil {
		return err
	}

	err = p.Stream(ctx, b.storedb, b.cache,
		func(ctx context.Context, entry types.DataContainer) error {
			return w.Queue(ctx, entry)

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
			return w.Queue(ctx, entry)
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

	// TODO wait for all transactions to complete before starting the builder
	time.Sleep(time.Second * 30)

	// Before we start the construction ensure all the new vertices have been index
	l.Infof("Reindexing graph following vertex ingest")
	err := b.graphdb.TriggerReindex(ctx, graphdb.VERTEX_ONLY)
	if err != nil {
		return fmt.Errorf("vertex reindexing: %w", err)
	}

	// Paths can have dependencies so must be built in sequence
	l.Info("Starting path construction")
	for label, p := range b.paths {
		l.Infof("Building path %s", label)

		err := b.buildPath(ctx, p)
		if err != nil {
			l.Errorf("building path %s: %v", label, err)
			continue
		}
	}

	// We've inserted more vertices via paths, reindex once again!
	l.Infof("Reindexing graph following path inserts")
	err = b.graphdb.TriggerReindex(ctx, graphdb.DEFAULT)
	if err != nil {
		return fmt.Errorf("path reindexing: %w", err)
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

	// All insertions are complete, the graph will now be available for query. Ensure it is reindex first
	l.Infof("Reindexing final graph")
	err = b.graphdb.TriggerReindex(ctx, graphdb.DEFAULT)
	if err != nil {
		return fmt.Errorf("final graph reindexing: %w", err)
	}

	l.Info("Completed edge construction")
	return nil
}
