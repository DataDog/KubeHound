package providers

import (
	"context"
	"fmt"

	"github.com/DataDog/KubeHound/pkg/collector"
	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/DataDog/KubeHound/pkg/ingestor"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/cache"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/graphdb"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/storedb"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
)

type ProvidersFactoryConfig struct {
	CacheProvider cache.CacheProvider
	StoreProvider storedb.Provider
	GraphProvider graphdb.Provider
}

// Initiating all the providers need for KubeHound (cache, store, graph)
func NewProvidersFactoryConfig(ctx context.Context, khCfg *config.KubehoundConfig) (*ProvidersFactoryConfig, error) {
	l := log.Logger(ctx)
	// Create the cache client
	l.Info("Loading cache provider")
	cp, err := cache.Factory(ctx, khCfg)
	if err != nil {
		return nil, fmt.Errorf("cache client creation: %w", err)
	}
	l.Info("Loaded cache provider", log.String("provider", cp.Name()))

	err = cp.Prepare(ctx)
	if err != nil {
		return nil, fmt.Errorf("store database prepare: %w", err)
	}

	// Create the store client
	l.Info("Loading store database provider")
	sp, err := storedb.Factory(ctx, khCfg)
	if err != nil {
		return nil, fmt.Errorf("store database client creation: %w", err)
	}
	msg := fmt.Sprintf("Loaded %s store provider", sp.Name())
	l.Info(msg, log.String("provider", sp.Name()))

	err = sp.Prepare(ctx)
	if err != nil {
		return nil, fmt.Errorf("store database prepare: %w", err)
	}

	// Create the graph client
	l.Info("Loading graph database provider")
	gp, err := graphdb.Factory(ctx, khCfg)
	if err != nil {
		return nil, fmt.Errorf("graph database client creation: %w", err)
	}
	msg = fmt.Sprintf("Loaded %s graph provider", gp.Name())
	l.Info(msg, log.String("provider", sp.Name()))

	err = gp.Prepare(ctx)
	if err != nil {
		return nil, fmt.Errorf("graph database prepare: %w", err)
	}

	return &ProvidersFactoryConfig{
		CacheProvider: cp,
		StoreProvider: sp,
		GraphProvider: gp,
	}, nil
}

func (p *ProvidersFactoryConfig) Close(ctx context.Context) {
	p.CacheProvider.Close(ctx)
	p.StoreProvider.Close(ctx)
	p.GraphProvider.Close(ctx)
}

func (p *ProvidersFactoryConfig) IngestBuildData(ctx context.Context, khCfg *config.KubehoundConfig) error {
	l := log.Logger(ctx)
	// Create the collector instance
	l.Info("Loading Kubernetes data collector client")
	collect, err := collector.ClientFactory(ctx, khCfg)
	if err != nil {
		return fmt.Errorf("collector client creation: %w", err)
	}
	defer func() { collect.Close(ctx) }()
	l.Infof("Loaded %s collector client", collect.Name())

	// Run the ingest pipeline
	l.Info("Starting Kubernetes raw data ingest")
	err = ingestor.IngestData(ctx, khCfg, collect, p.CacheProvider, p.StoreProvider, p.GraphProvider)
	if err != nil {
		return fmt.Errorf("raw data ingest: %w", err)
	}

	err = graph.BuildGraph(ctx, khCfg, p.StoreProvider, p.GraphProvider, p.CacheProvider)
	if err != nil {
		return err
	}

	return nil
}
