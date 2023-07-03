package core

import (
	"context"
	"fmt"

	"github.com/DataDog/KubeHound/pkg/collector"
	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/edge"
	"github.com/DataDog/KubeHound/pkg/kubehound/ingestor"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/cache"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/graphdb"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/storedb"
	"github.com/DataDog/KubeHound/pkg/telemetry"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	"github.com/google/uuid"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

func ingestData(ctx context.Context, cfg *config.KubehoundConfig, cache cache.CacheProvider,
	storedb storedb.Provider, graphdb graphdb.Provider) error {
	span, ctx := tracer.StartSpanFromContext(ctx, telemetry.SpanOperationIngestData, tracer.Measured())
	defer span.Finish()

	log.I.Info("Loading Kubernetes data collector client")
	collect, err := collector.ClientFactory(ctx, cfg)
	if err != nil {
		return fmt.Errorf("collector client creation: %w", err)
	}
	defer collect.Close(ctx)
	log.I.Infof("Loaded %s collector client", collect.Name())

	log.I.Info("Loading data ingestor")
	ingest, err := ingestor.Factory(cfg, collect, cache, storedb, graphdb)
	if err != nil {
		return fmt.Errorf("ingestor creation: %w", err)
	}
	defer ingest.Close(ctx)

	log.I.Info("Running dependency health checks")
	if err := ingest.HealthCheck(ctx); err != nil {
		return fmt.Errorf("ingestor dependency health check: %w", err)
	}

	log.I.Info("Running data ingest and normalization")
	if err := ingest.Run(ctx); err != nil {
		return fmt.Errorf("ingest: %w", err)
	}

	log.I.Info("Completed data ingest and normalization")
	return nil
}

// buildGraph will construct the attack graph by calculating and inserting all registered edges in parallel.
// All I/O operations are performed asynchronously.
func buildGraph(ctx context.Context, cfg *config.KubehoundConfig, storedb storedb.Provider,
	graphdb graphdb.Provider, cache cache.CacheReader) error {
	span, ctx := tracer.StartSpanFromContext(ctx, telemetry.SpanOperationBuildGraph, tracer.Measured())
	defer span.Finish()

	log.I.Info("Loading graph edge definitions")
	edges := edge.Registered()

	log.I.Info("Loading graph builder")
	builder, err := graph.NewBuilder(cfg, storedb, graphdb, cache, edges)
	if err != nil {
		return fmt.Errorf("graph builder creation: %w", err)
	}

	log.I.Info("Running dependency health checks")
	if err := builder.HealthCheck(ctx); err != nil {
		return fmt.Errorf("graph builder dependency health check: %w", err)
	}

	log.I.Info("Constructing graph")
	if err := builder.Run(ctx); err != nil {
		return fmt.Errorf("graph builder edge calculation: %w", err)
	}

	log.I.Info("Completed graph construction")
	return nil
}

// Launch will launch the KubeHound application to ingest data from a collector and create an attack graph.
func Launch(ctx context.Context, opts ...LaunchOption) error {
	runUUID := uuid.NewString()
	span, ctx := tracer.StartSpanFromContext(ctx, telemetry.SpanOperationLaunch, tracer.Measured())
	// We set this so we can measure run by run in addition of version per version
	// Useful when rerunning the same binary (same version) on different dataset or with different databases...
	span.SetBaggageItem("run_id", runUUID)
	// We update the base tags to include that run id, so we have it available for metrics
	tagRunUUID := "run_id:" + runUUID
	log.I.SetRunUUID(runUUID)
	telemetry.BaseTags = append(telemetry.BaseTags, tagRunUUID)
	defer span.Finish()

	log.I.Infof("Starting KubeHound (run_id: %s)", runUUID)
	log.I.Info("Initializing launch options")
	lOpts := &launchConfig{}
	for _, opt := range opts {
		opt(lOpts)
	}

	log.I.Info("Loading application configuration")
	var cfg *config.KubehoundConfig
	if len(lOpts.ConfigPath) != 0 {
		cfg = config.MustLoadConfig(lOpts.ConfigPath)
	} else {
		cfg = config.MustLoadDefaultConfig()
	}

	log.I.Info("Initializing application telemetry")
	err := telemetry.Initialize(cfg)
	if err != nil {
		log.I.Warnf("failed telemetry initialization: %v", err)
	}
	defer telemetry.Shutdown()

	log.I.Info("Loading cache provider")
	cp, err := cache.Factory(ctx, cfg)
	if err != nil {
		return fmt.Errorf("cache client creation: %w", err)
	}
	defer cp.Close(ctx)
	log.I.Infof("Loaded %s cache provider", cp.Name())

	log.I.Info("Loading store database provider")
	sp, err := storedb.Factory(ctx, cfg)
	if err != nil {
		return fmt.Errorf("store database client creation: %w", err)
	}
	defer sp.Close(ctx)
	log.I.Infof("Loaded %s store provider", sp.Name())

	log.I.Info("Loading graph database provider")
	gp, err := graphdb.Factory(ctx, cfg)
	if err != nil {
		return fmt.Errorf("graph database client creation: %w", err)
	}
	defer gp.Close(ctx)
	log.I.Infof("Loaded %s graph provider", gp.Name())

	log.I.Info("Starting Kubernetes raw data ingest")
	if err := ingestData(ctx, cfg, cp, sp, gp); err != nil {
		return fmt.Errorf("raw data ingest: %w", err)
	}

	log.I.Info("Building attack graph")
	if err := buildGraph(ctx, cfg, sp, gp, cp); err != nil {
		return fmt.Errorf("building attack graph: %w", err)
	}

	log.I.Info("Attack graph generation complete")
	return nil
}
