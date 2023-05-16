package core

import (
	"context"
	"fmt"

	"github.com/DataDog/KubeHound/pkg/collector"
	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/DataDog/KubeHound/pkg/telemetry"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	"github.com/DataDog/ase-experimental/kubehound/kh/graph/edge"
	"github.com/DataDog/ase-experimental/kubehound/kh/graphbuilder"
	"github.com/DataDog/ase-experimental/kubehound/kh/ingestor"
	"github.com/DataDog/ase-experimental/kubehound/kh/storage/cache"
	"github.com/DataDog/ase-experimental/kubehound/kh/storage/graphdb"
	"github.com/DataDog/ase-experimental/kubehound/kh/storage/storedb"
)

// ingestData will stream data from the collector through the ingestion pipeline.
// Each object ingested will be processed, write one or more entries to the intermediate store and cache,
// then create one or more vertices in the graph database. All I/O operations are performed asynchronously.
func ingestData(ctx context.Context, cfg *config.KubehoundConfig, cache cache.CacheProvider,
	storedb storedb.Provider, graphdb graphdb.Provider) error {

	log.I.Info("Loading Kubernetes data collector client")
	collect, err := collector.ClientFactory(ctx, cfg)
	if err != nil {
		return fmt.Errorf("collector client creation: %w", err)
	}
	defer collect.Close(ctx)

	log.I.Info("Loading data ingestor")
	ingest, err := ingestor.New(cfg, collect, cache, storedb, graphdb)
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
	graphdb graphdb.Provider) error {

	log.I.Info("Loading graph edge definitions")
	edges := edge.Registry()

	log.I.Info("Loading graph builder")
	builder, err := graphbuilder.New(cfg, storedb, graphdb, edges)
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
func Launch(ctx context.Context) error {
	log.I.Info("Initializing application telemetry")
	tc, err := telemetry.Initialize()
	if err != nil {
		log.I.Warnf("failed telemetry initialization: %v", err)
	}
	defer tc.Shutdown()

	log.I.Info("Loading application configuration")
	cfg := config.MustLoadDefaultConfig()

	log.I.Info("Loading cache provider")
	cp, err := cache.Factory(cfg)
	if err != nil {
		return fmt.Errorf("cache client creation: %w", err)
	}
	defer cp.Close(ctx)

	log.I.Info("Loading store database provider")
	sp, err := storedb.Factory(cfg)
	if err != nil {
		return fmt.Errorf("store database client creation: %w", err)
	}
	defer sp.Close(ctx)

	log.I.Info("Loading graph database provider")
	gp, err := graphdb.Factory(cfg)
	if err != nil {
		return fmt.Errorf("graph database client creation: %w", err)
	}
	defer gp.Close(ctx)

	log.I.Info("Starting Kubernetes raw data ingest")
	if err := ingestData(ctx, cfg, cp, sp, gp); err != nil {
		return fmt.Errorf("raw data ingest: %w", err)
	}

	log.I.Info("Building attack graph")
	if err := buildGraph(ctx, cfg, sp, gp); err != nil {
		return fmt.Errorf("building attack graph: %w", err)
	}

	log.I.Info("Attack graph generation complete")
	return nil
}
