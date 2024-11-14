package core

import (
	"context"
	"fmt"

	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/DataDog/KubeHound/pkg/ingestor/api"
	"github.com/DataDog/KubeHound/pkg/ingestor/api/grpc"
	"github.com/DataDog/KubeHound/pkg/ingestor/notifier/noop"
	"github.com/DataDog/KubeHound/pkg/ingestor/puller/blob"
	"github.com/DataDog/KubeHound/pkg/kubehound/providers"
	"github.com/DataDog/KubeHound/pkg/telemetry/events"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	"github.com/DataDog/KubeHound/pkg/telemetry/span"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

func initCoreGrpcApi(ctx context.Context, khCfg *config.KubehoundConfig) (*api.IngestorAPI, error) {
	l := log.Logger(ctx)
	l.Info("Starting KubeHound Distributed Ingestor Service")
	span, ctx := span.SpanRunFromContext(ctx, span.IngestorLaunch)
	var err error
	defer func() {
		span.Finish(tracer.WithError(err))
	}()

	// Initialize the providers (graph, cache, store)
	l.Info("Initializing providers (graph, cache, store)")
	p, err := providers.NewProvidersFactoryConfig(ctx, khCfg)
	if err != nil {
		return nil, fmt.Errorf("factory config creation: %w", err)
	}

	l.Info("Creating Blob Storage provider")
	puller, err := blob.NewBlobStorage(khCfg, khCfg.Ingestor.Blob)
	if err != nil {
		return nil, err
	}

	l.Info("Creating Noop Notifier")
	noopNotifier := noop.NewNoopNotifier()

	l.Info("Creating Ingestor API")

	return api.NewIngestorAPI(khCfg, puller, noopNotifier, p), nil
}

func CoreGrpcApi(ctx context.Context, khCfg *config.KubehoundConfig) error {
	ingestorApi, err := initCoreGrpcApi(ctx, khCfg)
	if err != nil {
		_ = events.PushEvent(ctx, events.IngestorFailed, "")

		return err
	}
	defer ingestorApi.Close(ctx)
	_ = events.PushEvent(ctx, events.IngestorInit, "")

	l := log.Logger(ctx)
	l.Info("Starting Ingestor API")
	err = grpc.Listen(ctx, ingestorApi)
	if err != nil {
		return err
	}

	l.Info("KubeHound Ingestor API shutdown")

	return nil
}
