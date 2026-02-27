package graph

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/edge"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/vertex"
	"github.com/DataDog/KubeHound/pkg/kubehound/services"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/graphdb"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/storedb"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	"github.com/DataDog/KubeHound/pkg/telemetry/span"
	"github.com/DataDog/KubeHound/pkg/telemetry/tag"
	"github.com/DataDog/KubeHound/pkg/worker"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

// Builder handles the construction of the graph vertices and edges.
type Builder struct {
	cfg     *config.KubehoundConfig
	storedb storedb.Provider
	graphdb graphdb.Provider
	db      *sql.DB
	edges   *edge.Registry
}

// NewBuilder returns a new builder instance from the provided application config and service dependencies.
func NewBuilder(cfg *config.KubehoundConfig, store storedb.Provider, graph graphdb.Provider,
	db *sql.DB, edges *edge.Registry) (*Builder, error) {
	n := &Builder{
		cfg:     cfg,
		storedb: store,
		graphdb: graph,
		db:      db,
		edges:   edges,
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

// buildVertex inserts all vertices of a single type into the graph database.
func (b *Builder) buildVertex(ctx context.Context, vb vertex.Builder) error {
	span, ctx := span.StartSpanFromContext(ctx, span.BuildEdge, tracer.Measured(), tracer.ResourceName(vb.Label()))
	span.SetTag(tag.LabelTag, vb.Label())
	var err error
	defer func() { span.Finish(tracer.WithError(err)) }()

	l := log.Logger(ctx)
	l.Info("Building vertex", log.String("label", vb.Label()))

	if err = vb.Initialize(b.cfg); err != nil {
		return err
	}

	w, err := b.graphdb.VertexWriter(ctx, vb, b.db, graphdb.WithTags(tag.GetBaseTags()))
	if err != nil {
		return err
	}
	defer w.Close(ctx)

	rows, err := b.db.QueryContext(ctx, vb.Query(b.cfg.Dynamic.RunID.String(), b.cfg.Dynamic.Cluster.Name))
	if err != nil {
		return fmt.Errorf("vertex query %s: %w", vb.Label(), err)
	}
	defer rows.Close()

	for rows.Next() {
		m, err := vb.Scanner(rows)
		if err != nil {
			return fmt.Errorf("vertex scan %s: %w", vb.Label(), err)
		}
		if err := w.Queue(ctx, m); err != nil {
			return fmt.Errorf("vertex queue %s: %w", vb.Label(), err)
		}
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("vertex rows %s: %w", vb.Label(), err)
	}

	return w.Flush(ctx)
}

// buildVertices inserts all vertex types into the graph database.
func (b *Builder) buildVertices(ctx context.Context) error {
	vertexBuilders := []vertex.Builder{
		&vertex.Node{},
		&vertex.Pod{},
		&vertex.Container{},
		&vertex.Volume{},
		&vertex.Identity{},
		&vertex.PermissionSet{},
		&vertex.Endpoint{},
	}
	for _, vb := range vertexBuilders {
		if err := b.buildVertex(ctx, vb); err != nil {
			return fmt.Errorf("building vertex %s: %w", vb.Label(), err)
		}
	}
	return nil
}

// buildEdge inserts a class of edges into the graph database.
func (b *Builder) buildEdge(ctx context.Context, label string, e edge.Builder) error {
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
	defer w.Close(ctx)

	return e.Stream(ctx, b.db, w)
}

// buildMutating constructs all the mutating edges in the graph database.
func (b *Builder) buildMutating(ctx context.Context) error {
	l := log.Logger(ctx)
	for label, e := range b.edges.Mutating() {
		err := b.buildEdge(ctx, label, e)
		if err != nil {
			if b.cfg.Builder.StopOnError {
				return fmt.Errorf("building mutating edge %s: %w", label, err)
			}
			l.Warnf("Failed to create a mutating edge (type: %s). The created graph will be INCOMPLETE (change `builder.stop_on_error` to abort or error instead): %v", e.Name(), err)

			return nil
		}
	}

	return nil
}

// buildSimple constructs all the simple edges in the graph database.
func (b *Builder) buildSimple(ctx context.Context) error {
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
		wp.Submit(func() error {
			err := b.buildEdge(workCtx, label, e)
			if err != nil {
				if b.cfg.Builder.StopOnError {
					return err
				}
				l.Warnf("Failed to create a simple edge (type: %s). The created graph will be INCOMPLETE (change `builder.stop_on_error` to abort or error instead): %v", e.Name(), err)

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
func (b *Builder) buildDependent(ctx context.Context) error {
	l := log.Logger(ctx)
	for label, e := range b.edges.Dependent() {
		err := b.buildEdge(ctx, label, e)
		if err != nil {
			if b.cfg.Builder.StopOnError {
				return fmt.Errorf("building dependent edge %s: %w", label, err)
			}
			l.Warnf("Failed to create a dependent edge (type: %s). The created graph will be INCOMPLETE (change `builder.stop_on_error` to abort or error instead): %v", e.Name(), err)

			return nil
		}
	}

	return nil
}

// Run constructs all vertices and edges in the graph database.
func (b *Builder) Run(ctx context.Context) error {
	l := log.Trace(ctx)

	if b.cfg.Builder.Edge.LargeClusterOptimizations {
		log.Trace(ctx).Warnf("Using large cluster optimizations in graph construction")
	}

	// Phase 1: Insert all vertices
	l.Info("Starting vertex construction")
	if err := b.buildVertices(ctx); err != nil {
		return err
	}

	// Phase 2: Build edges
	// Mutating edges must be built first, sequentially
	l.Info("Starting mutating edge construction")
	if err := b.buildMutating(ctx); err != nil {
		return err
	}

	// Simple edges can be built in parallel
	l.Info("Starting simple edge construction")
	if err := b.buildSimple(ctx); err != nil {
		return err
	}

	// Dependent edges must be built last, sequentially
	l.Info("Starting dependent edge construction")
	if err := b.buildDependent(ctx); err != nil {
		return err
	}

	l.Info("Completed graph construction")

	return nil
}

// BuildGraph will construct the attack graph by calculating and inserting all registered vertices and edges.
func BuildGraph(outer context.Context, cfg *config.KubehoundConfig, storedb storedb.Provider,
	graphdb graphdb.Provider) error {
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

	db, ok := storedb.Reader().(*sql.DB)
	if !ok {
		return fmt.Errorf("store provider reader is not *sql.DB")
	}

	l.Info("Loading graph builder")
	builder, err := NewBuilder(cfg, storedb, graphdb, db, edges)
	if err != nil {
		return fmt.Errorf("graph builder creation: %w", err)
	}

	l.Info("Running dependency health checks")
	if err := builder.HealthCheck(ctx); err != nil {
		return fmt.Errorf("graph builder dependency health check: %w", err)
	}

	l.Info("Constructing graph")
	if err := builder.Run(ctx); err != nil {
		return fmt.Errorf("graph builder construction: %w", err)
	}

	l.Info("Completed graph construction", log.Duration("duration", time.Since(start)))

	return nil
}
