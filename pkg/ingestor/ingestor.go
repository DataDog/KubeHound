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
	"github.com/DataDog/KubeHound/pkg/telemetry/span"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

func IngestData(ctx context.Context, cfg *config.KubehoundConfig, collect collector.CollectorClient, cache cache.CacheProvider,
	storedb storedb.Provider, graphdb graphdb.Provider) error {

	start := time.Now()
	_ = start
	span, ctx := tracer.StartSpanFromContext(ctx, span.IngestData, tracer.Measured())
	var err error
	defer func() { span.Finish(tracer.WithError(err)) }()

	//log.I..Info("Loading data ingestor")
	ingest, err := ingestor.Factory(cfg, collect, cache, storedb, graphdb)
	if err != nil {
		return fmt.Errorf("ingestor creation: %w", err)
	}
	defer ingest.Close(ctx)

	//log.I..Info("Running dependency health checks")
	if err := ingest.HealthCheck(ctx); err != nil {
		return fmt.Errorf("ingestor dependency health check: %w", err)
	}

	//log.I..Info("Running data ingest and normalization")
	if err := ingest.Run(ctx); err != nil {
		return fmt.Errorf("ingest: %w", err)
	}

	//log.I..Infof("Completed data ingest and normalization in %s", time.Since(start))

	return nil
}
