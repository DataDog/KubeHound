package ingestor

import (
	"context"
	"fmt"
	"time"

	"github.com/DataDog/KubeHound/pkg/collector"
	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/DataDog/KubeHound/pkg/kubehound/ingestor"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/cache"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/graphdb"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/storedb"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	"github.com/DataDog/KubeHound/pkg/telemetry/span"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

func IngestData(ctx context.Context, cfg *config.KubehoundConfig, collect collector.CollectorClient, cache cache.CacheProvider,
	storedb storedb.Provider, graphdb graphdb.Provider) error {
	l := log.Logger(ctx)

	start := time.Now()
	span, ctx := span.SpanRunFromContext(ctx, span.IngestData)
	var err error
	defer func() { span.Finish(tracer.WithError(err)) }()

	l.Info("Loading data ingestor")
	ingest, err := ingestor.Factory(cfg, collect, cache, storedb, graphdb)
	if err != nil {
		return fmt.Errorf("ingestor creation: %w", err)
	}
	defer ingest.Close(ctx)

	l.Info("Running dependency health checks")
	if err := ingest.HealthCheck(ctx); err != nil {
		return fmt.Errorf("ingestor dependency health check: %w", err)
	}

	l.Info("Running data ingest and normalization")
	if err := ingest.Run(ctx); err != nil {
		return fmt.Errorf("ingest: %w", err)
	}

	l.Info("Completed data ingest and normalization", log.Duration("time", time.Since(start)))

	return nil
}
