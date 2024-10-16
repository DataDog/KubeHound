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
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	"github.com/DataDog/KubeHound/pkg/telemetry/span"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

func CoreGrpcApi(ctx context.Context, khCfg *config.KubehoundConfig) error {
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
		return fmt.Errorf("factory config creation: %w", err)
	}
	defer p.Close(ctx)

	l.Info("Creating Blob Storage provider")
	puller, err := blob.NewBlobStorage(khCfg, khCfg.Ingestor.Blob)
	if err != nil {
		return err
	}

	l.Info("Creating Noop Notifier")
	noopNotifier := noop.NewNoopNotifier()

	l.Info("Creating Ingestor API")
	ingestorApi := api.NewIngestorAPI(khCfg, puller, noopNotifier, p)

	l.Info("Starting Ingestor API")
	err = grpc.Listen(ctx, ingestorApi)
	if err != nil {
		return err
	}

	l.Info("KubeHound Ingestor API shutdown")

	return nil
}
