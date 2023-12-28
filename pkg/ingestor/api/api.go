package api

import (
	"context"

	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/DataDog/KubeHound/pkg/ingestor/notifier"
	"github.com/DataDog/KubeHound/pkg/ingestor/puller"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/cache"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/graphdb"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/storedb"
)

type API interface {
	Ingest(ctx context.Context, clusterName string, runID string) error
	Notify(ctx context.Context, clusterName string, runID string) error
}

//go:generate protoc --go_out=./pb --go_opt=paths=source_relative --go-grpc_out=./pb --go-grpc_opt=paths=source_relative api.proto
type IngestorAPI struct {
	puller   puller.DataPuller
	notifier notifier.Notifier
	Cfg      *config.KubehoundConfig
	storedb  storedb.Provider
	graphdb  graphdb.Provider
	cache    cache.CacheReader
}

var _ API = (*IngestorAPI)(nil)

func NewIngestorAPI(cfg *config.KubehoundConfig, puller puller.DataPuller, notifier notifier.Notifier,
	sdb storedb.Provider, gdb graphdb.Provider, c cache.CacheReader) *IngestorAPI {
	return &IngestorAPI{
		notifier: notifier,
		puller:   puller,
		Cfg:      cfg,
		storedb:  sdb,
		graphdb:  gdb,
		cache:    c,
	}
}

func (g *IngestorAPI) Ingest(ctx context.Context, clusterName string, runID string) error {
	archivePath, err := g.puller.Pull(ctx, clusterName, runID)
	if err != nil {
		return err
	}
	err = g.puller.Close(ctx, archivePath)
	if err != nil {
		return err
	}
	err = g.puller.Extract(ctx, archivePath)
	if err != nil {
		return err
	}
	err = graph.BuildGraph(ctx, g.Cfg, g.storedb, g.graphdb, g.cache)
	if err != nil {
		return err
	}
	err = g.notifier.Notify(ctx, clusterName, runID)
	if err != nil {
		return err
	}

	return nil
}

// Notify notifies the caller that the ingestion is completed
func (g *IngestorAPI) Notify(ctx context.Context, clusterName string, runID string) error {
	return g.notifier.Notify(ctx, clusterName, runID)
}
