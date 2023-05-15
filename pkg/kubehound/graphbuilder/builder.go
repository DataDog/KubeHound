package graphbuilder

import (
	"context"
	"fmt"

	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/DataDog/KubeHound/pkg/globals"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/edge"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/graphdb"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/storedb"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
)

const (
	ComponentName = "graph-builder"
)

type Builder struct {
	cfg      *config.KubehoundConfig
	storedb  storedb.Provider
	graphdb  graphdb.Provider
	registry edge.EdgeRegistry
}

func New(cfg *config.KubehoundConfig, store storedb.Provider,
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

func (b *Builder) Run(ctx context.Context) error {
	l := log.Trace(ctx, log.WithComponent(ComponentName))
	wp, err := NewWorkerPool(b.cfg)
	if err != nil {
		return fmt.Errorf("graph builder worker pool create: %w", err)
	}

	l.Info("Starting edge construction")
	for label, e := range edge.Registry() {
		wp.Submit(func() {
			l.Infof("Building edge %s", label)

			err := b.buildEdge(ctx, e)
			if err != nil {
				l.Errorf("error building edge %s: %w", label, err)
			}
		})
	}

	wp.WaitForComplete()
	l.Info("Completed edge construction")

	return nil
}
