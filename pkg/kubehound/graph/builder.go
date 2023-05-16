package graph

import (
	"context"
	"fmt"

	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/DataDog/KubeHound/pkg/globals"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/edge"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/graphdb"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/storedb"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	"github.com/DataDog/KubeHound/pkg/worker"
)

const (
	BuilderComponentName = "graph-builder"
)

// Builder handles the construction of the graph edges once vertices have been ingested via the ingestion pipeline.
type Builder struct {
	cfg      *config.KubehoundConfig
	storedb  storedb.Provider
	graphdb  graphdb.Provider
	registry edge.EdgeRegistry
}

// NewBuilder returns a new builder instance from the provided application config and service dependencies.
func NewBuilder(cfg *config.KubehoundConfig, store storedb.Provider,
	graph graphdb.Provider, registry edge.EdgeRegistry) (*Builder, error) {

	n := &Builder{
		cfg:      cfg,
		storedb:  store,
		graphdb:  graph,
		registry: registry,
	}

	return n, nil
}

// HealthCheck provides a mechanism for the caller to check health of the builder dependencies.
func (b *Builder) HealthCheck(ctx context.Context) error {
	return globals.ErrNotImplemented
}

// buildEdge inserts a class of edges into the graph database.
// NOTE: function is blocking and expected to be called from within a goroutine.
func (b *Builder) buildEdge(ctx context.Context, e edge.Edge) error {
	w, err := b.graphdb.EdgeWriter(ctx, e.Traversal())
	if err != nil {
		return err
	}

	err = e.Stream(ctx, b.storedb,
		func(ctx context.Context, entry edge.DataContainer) error {
			processed, err := e.Processor(ctx, entry)
			// TODO option for skip write if signalled by processor

			if err != nil {
				// TODO tolerate errors
				return err
			}

			return w.Queue(ctx, processed)

		},
		func(ctx context.Context) error {
			complete, err := w.Flush(ctx)
			if err != nil {
				return err
			}

			select {
			case <-complete:
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		})

	w.Close(ctx)

	return err
}

// runInternal implements the Run function and is extracted separately to enable testing.
func (b *Builder) runInternal(outer context.Context, registry edge.EdgeRegistry) error {
	var err error
	ctx, cancel := context.WithCancelCause(outer)
	defer cancel(err)

	l := log.Trace(ctx, log.WithComponent(BuilderComponentName))
	l.Info("Creating edge builder worker pool")
	wp, err := worker.PoolFactory(b.cfg)
	if err != nil {
		return fmt.Errorf("graph builder worker pool create: %w", err)
	}

	if err := wp.Start(ctx); err != nil {
		return fmt.Errorf("graph builder worker pool start: %w", err)
	}

	l.Info("Starting edge construction")
	for label, e := range registry {
		e := e
		label := label

		err := wp.Submit(func() {
			l.Infof("Building edge %s", label)

			err := b.buildEdge(ctx, e)
			if err != nil {
				l.Errorf("building edge %s: %w", label, err)
				cancel(err) // Ensure all work is stopped on an edge error

			}
		})
		if err != nil {
			l.Errorf("submitting edge %s to worker pool: %w", label, err)
			cancel(err) // Ensure all work is stopped on any submit error
		}
	}

	wp.WaitForComplete()
	l.Info("Completed edge construction")

	return nil
}

// Run constructs all the registered edges in the graph database.
// NOTE: edges are constructed in parallel using a worker pool with properties configured via the top-level KubeHound config.
func (b *Builder) Run(ctx context.Context) error {
	return b.runInternal(ctx, edge.Registry())
}
