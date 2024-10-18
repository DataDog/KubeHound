package core

import (
	"context"
	"fmt"
	"time"

	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/DataDog/KubeHound/pkg/kubehound/providers"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	"github.com/DataDog/KubeHound/pkg/telemetry/span"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

// Setting the current cluster targeted for the live run.
func CoreInitLive(ctx context.Context, khCfg *config.KubehoundConfig) error {
	clusterName, err := config.GetClusterName(ctx)
	if err != nil {
		return fmt.Errorf("collector cluster info: %w", err)
	}
	khCfg.Dynamic.ClusterName = clusterName

	return nil
}

// CoreLive will launch the KubeHound application to ingest data from a collector and create an attack graph.
func CoreLive(ctx context.Context, khCfg *config.KubehoundConfig) error {
	l := log.Logger(ctx)
	span, ctx := span.SpanRunFromContext(ctx, span.Launch)
	var err error
	defer func() { span.Finish(tracer.WithError(err)) }()

	// Check for run configuration
	err = khCfg.Dynamic.HealthCheck()
	if err != nil {
		return fmt.Errorf("health check: %w", err)
	}

	// Start the run
	start := time.Now()
	l.Info("Starting KubeHound", log.String(log.FieldRunIDKey, khCfg.Dynamic.RunID.String()), log.String("cluster_name", khCfg.Dynamic.ClusterName))

	// Initialize the providers (graph, cache, store)
	l.Info("Initializing providers (graph, cache, store)")
	p, err := providers.NewProvidersFactoryConfig(ctx, khCfg)
	if err != nil {
		return fmt.Errorf("factory config creation: %w", err)
	}
	defer p.Close(ctx)

	// Running the ingestion pipeline (ingestion and building the graph)
	l.Info("Running the ingestion pipeline")
	err = p.IngestBuildData(ctx, khCfg)
	if err != nil {
		return fmt.Errorf("ingest build data: %w", err)
	}

	l.Info("KubeHound run complete", log.String(log.FieldRunIDKey, khCfg.Dynamic.RunID.String()), log.Duration("duration", time.Since(start)))

	return nil
}
