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

type Builder struct {
	cfg      *config.KubehoundConfig
	storedb  storedb.Provider
	graphdb  graphdb.Provider
	registry edge.EdgeRegistry
}

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

func (b *Builder) HealthCheck(ctx context.Context) error {
	return globals.ErrNotImplemented
}

func (b *Builder) buildEdge(ctx context.Context, e edge.Edge) error {
	w, err := b.graphdb.EdgeWriter(ctx, e.Traversal())
	if err != nil {
		return err
	}

	// l := log.Trace(ctx, log.WithComponent(ComponentName))
	err = e.Stream(ctx, b.storedb,
		func(ctx context.Context, entry interface{}) error {
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

			<-complete
			return nil
		})

	w.Close(ctx)
	if err != nil {
		return err
	}

	return nil
}

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

	// Ensure all work is stopped on an edge error
	go func() {
		<-ctx.Done()
		if context.Cause(ctx) != nil {
			wp.Stop()
		}
	}()

	l.Info("Starting edge construction")
	for label, e := range registry {
		e := e
		label := label

		err := wp.Submit(func() {
			l.Infof("Building edge %s", label)

			err := b.buildEdge(ctx, e)
			if err != nil {
				l.Errorf("building edge %s: %w", label, err)
				cancel(err)
			}
		})
		if err != nil {
			l.Errorf("submitting edge %s to worker pool: %w", label, err)
			cancel(err)
		}
	}

	wp.WaitForComplete()
	l.Info("Completed edge construction")

	return nil
}

func (b *Builder) Run(ctx context.Context) error {
	return b.runInternal(ctx, edge.Registry())
}
