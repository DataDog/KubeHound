package ingestor

import (
	"context"

	"github.com/DataDog/KubeHound/pkg/collector"
	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/DataDog/KubeHound/pkg/globals"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/cache"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/graphdb"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/storedb"
)

type Ingestor interface {
	// HealthCheck provides a mechanism for the client to check health of the provider.
	HealthCheck(ctx context.Context) error

	// Run starts the ingestion pipelines to ingest data provided by the collector into the intermediate store and graph database.
	Run(ctx context.Context) error

	// Close cleans up any resources used by the Provider implementation. Provider cannot be reused after this call.
	Close(ctx context.Context) error
}

// Factory creates a new ingestor instance from the provided configuration and service dependencies.
func Factory(cfg *config.KubehoundConfig, collect collector.CollectorClient, cache cache.CacheProvider,
	storedb storedb.Provider, graphdb graphdb.Provider) (Ingestor, error) {

	return nil, globals.ErrNotImplemented
}
