package ingestor

import (
	"context"
	"fmt"
	"time"

	"github.com/DataDog/KubeHound/pkg/collector"
	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/DataDog/KubeHound/pkg/kubehound/ingestor"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/storedb"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	"github.com/DataDog/KubeHound/pkg/telemetry/span"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

func IngestData(ctx context.Context, cfg *config.KubehoundConfig, collect collector.CollectorClient,
	storedb storedb.Provider) error {
	l := log.Logger(ctx)

	start := time.Now()
	span, ctx := span.SpanRunFromContext(ctx, span.IngestData)
	var err error
	defer func() { span.Finish(tracer.WithError(err)) }()

	l.Info("Running data ingest and normalization")
	if err = ingestor.Collect(ctx, cfg, collect, storedb); err != nil {
		return fmt.Errorf("ingest: %w", err)
	}

	l.Info("Completed data ingest and normalization", log.Duration("time", time.Since(start)))

	return nil
}
