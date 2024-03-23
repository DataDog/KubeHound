package api

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/DataDog/KubeHound/pkg/collector"
	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/DataDog/KubeHound/pkg/ingestor"
	"github.com/DataDog/KubeHound/pkg/ingestor/notifier"
	"github.com/DataDog/KubeHound/pkg/ingestor/puller"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph"
	"github.com/DataDog/KubeHound/pkg/kubehound/providers"
	"github.com/DataDog/KubeHound/pkg/telemetry/events"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	"github.com/DataDog/KubeHound/pkg/telemetry/span"
	"github.com/DataDog/KubeHound/pkg/telemetry/tag"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

type API interface {
	Ingest(ctx context.Context, clusterName string, runID string) error
	Notify(ctx context.Context, clusterName string, runID string) error
}

//go:generate protoc --go_out=./pb --go_opt=paths=source_relative --go-grpc_out=./pb --go-grpc_opt=paths=source_relative api.proto
type IngestorAPI struct {
	puller    puller.DataPuller
	notifier  notifier.Notifier
	Cfg       *config.KubehoundConfig
	providers *providers.ProvidersFactoryConfig
}

var _ API = (*IngestorAPI)(nil)

func NewIngestorAPI(cfg *config.KubehoundConfig, puller puller.DataPuller, notifier notifier.Notifier,
	p *providers.ProvidersFactoryConfig) *IngestorAPI {
	return &IngestorAPI{
		notifier:  notifier,
		puller:    puller,
		Cfg:       cfg,
		providers: p,
	}
}

func (g *IngestorAPI) Ingest(_ context.Context, clusterName string, runID string) error {
	events.PushEvent(
		fmt.Sprintf("Ingesting cluster %s with runID %s", clusterName, runID),
		fmt.Sprintf("Ingesting cluster %s with runID %s", clusterName, runID),
		[]string{
			tag.IngestionRunID(runID),
		},
	)
	// Settings global variables for the run in the context to propagate them to the spans
	runCtx := context.Background()
	runCtx = context.WithValue(runCtx, span.ContextLogFieldClusterName, clusterName)
	runCtx = context.WithValue(runCtx, span.ContextLogFieldRunID, runID)

	spanJob, runCtx := span.SpanIngestRunFromContext(runCtx, span.IngestorStartJob)
	var err error
	defer func() { spanJob.Finish(tracer.WithError(err)) }()

	archivePath, err := g.puller.Pull(runCtx, clusterName, runID) //nolint: contextcheck
	if err != nil {
		return err
	}
	err = g.puller.Close(runCtx, archivePath) //nolint: contextcheck
	if err != nil {
		return err
	}
	err = g.puller.Extract(runCtx, archivePath) //nolint: contextcheck
	if err != nil {
		return err
	}

	runCfg := g.Cfg
	runCfg.Collector = config.CollectorConfig{
		Type: config.CollectorTypeFile,
		File: &config.FileCollectorConfig{
			Directory:   filepath.Dir(archivePath),
			ClusterName: clusterName,
		},
	}

	// Create the collector instance
	log.I.Info("Loading Kubernetes data collector client")
	collect, err := collector.ClientFactory(runCtx, runCfg) //nolint: contextcheck
	if err != nil {
		return fmt.Errorf("collector client creation: %w", err)
	}
	defer collect.Close(runCtx) //nolint: contextcheck
	log.I.Infof("Loaded %s collector client", collect.Name())

	err = g.Cfg.ComputeDynamic(config.WithClusterName(clusterName), config.WithRunID(runID))
	if err != nil {
		return err
	}

	// Run the ingest pipeline
	log.I.Info("Starting Kubernetes raw data ingest")
	err = ingestor.IngestData(runCtx, runCfg, collect, g.providers.CacheProvider, g.providers.StoreProvider, g.providers.GraphProvider) //nolint: contextcheck
	if err != nil {
		return fmt.Errorf("raw data ingest: %w", err)
	}

	err = graph.BuildGraph(runCtx, runCfg, g.providers.StoreProvider, g.providers.GraphProvider, g.providers.CacheProvider) //nolint: contextcheck
	if err != nil {
		return err
	}
	err = g.notifier.Notify(runCtx, clusterName, runID) //nolint: contextcheck
	if err != nil {
		return err
	}

	return nil
}

// Notify notifies the caller that the ingestion is completed
func (g *IngestorAPI) Notify(ctx context.Context, clusterName string, runID string) error {
	return g.notifier.Notify(ctx, clusterName, runID)
}
