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

// CoreLive will launch the KubeHound application to ingest data from a collector and create an attack graph.
func CoreLive(ctx context.Context, khCfg *config.KubehoundConfig) error {
	span, ctx := tracer.StartSpanFromContext(ctx, span.Launch, tracer.Measured())
	var err error
	defer func() { span.Finish(tracer.WithError(err)) }()

	// Start the run
	start := time.Now()
	log.I.Infof("Starting KubeHound (run_id: %s)", khCfg.Dynamic.RunID.String())

	// Initialize the providers (graph, cache, store)
	log.I.Info("Initializing providers (graph, cache, store)")
	p, err := providers.NewProvidersFactoryConfig(ctx, khCfg)
	if err != nil {
		return fmt.Errorf("factory config creation: %w", err)
	}
	defer p.Close(ctx)

	// Running the ingestion pipeline (ingestion and building the graph)
	log.I.Info("Running the ingestion pipeline")
	err = p.IngestBuildData(ctx, khCfg)
	if err != nil {
		return fmt.Errorf("ingest build data: %w", err)
	}

	log.I.Infof("KubeHound run (id=%s) complete in %s", khCfg.Dynamic.RunID.String(), time.Since(start))

	return nil
}
