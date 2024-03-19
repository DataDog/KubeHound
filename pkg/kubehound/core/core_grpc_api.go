package core

import (
	"context"
	"fmt"

	"github.com/DataDog/KubeHound/pkg/ingestor/api"
	"github.com/DataDog/KubeHound/pkg/ingestor/api/grpc"
	"github.com/DataDog/KubeHound/pkg/ingestor/notifier/noop"
	"github.com/DataDog/KubeHound/pkg/ingestor/puller/blob"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	"github.com/DataDog/KubeHound/pkg/telemetry/span"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

func LaunchRemoteIngestor(ctx context.Context, opts ...LaunchOption) error {
	log.I.Infof("Starting KubeHound Distributed Ingestor Service")
	lOpts := &launchConfig{}
	for _, opt := range opts {
		opt(lOpts)
	}
	var err error

	ctx, lc := NewLaunchConfig(ctx, lOpts.ConfigPath, false)
	span, ctx := tracer.StartSpanFromContext(ctx, span.IngestorLaunch, tracer.Measured())
	err = lc.Initialize(ctx, true)
	if err != nil {
		return fmt.Errorf("initialize config: %w", err)
	}
	defer func() {
		span.Finish(tracer.WithError(err))
		lc.Close()
	}()

	fc, err := NewProvidersFactoryConfig(ctx, lc)
	if err != nil {
		return fmt.Errorf("factory config creation: %w", err)
	}
	defer fc.Close(ctx)

	log.I.Info("Creating Blob Storage provider")
	puller, err := blob.NewBlobStoragePuller(lc.Cfg)
	if err != nil {
		return err
	}

	log.I.Info("Creating Noop Notifier")
	noopNotifier := noop.NewNoopNotifier()

	log.I.Info("Creating Ingestor API")
	ingestorApi := api.NewIngestorAPI(lc.Cfg, puller, noopNotifier, fc.StoreProvider, fc.GraphProvider, fc.CacheProvider)

	log.I.Info("Starting Ingestor API")
	err = grpc.Listen(ctx, ingestorApi)
	if err != nil {
		return err
	}

	log.I.Infof("KubeHound Ingestor API shutdown")

	return nil
}
