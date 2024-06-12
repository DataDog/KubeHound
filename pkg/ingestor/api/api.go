package api

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/DataDog/KubeHound/pkg/collector"
	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/DataDog/KubeHound/pkg/ingestor"
	"github.com/DataDog/KubeHound/pkg/ingestor/notifier"
	"github.com/DataDog/KubeHound/pkg/ingestor/puller"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/adapter"
	"github.com/DataDog/KubeHound/pkg/kubehound/providers"
	"github.com/DataDog/KubeHound/pkg/kubehound/store/collections"
	"github.com/DataDog/KubeHound/pkg/telemetry/events"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	"github.com/DataDog/KubeHound/pkg/telemetry/span"
	"github.com/DataDog/KubeHound/pkg/telemetry/tag"
	gremlingo "github.com/apache/tinkerpop/gremlin-go/v3/driver"
	"go.mongodb.org/mongo-driver/bson"
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

var (
	_                  API = (*IngestorAPI)(nil)
	ErrAlreadyIngested     = errors.New("ingestion already completed")
)

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

	alreadyIngested, err := g.isAlreadyIngestedInGraph(runCtx, clusterName, runID) //nolint: contextcheck
	if err != nil {
		return err
	}

	if alreadyIngested {
		return fmt.Errorf("%w [%s:%s]", ErrAlreadyIngested, clusterName, runID)
	}

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

	// We need to flush the cache to prevent warnings/errors when overwriting elements in cache from the previous ingestion
	// This avoid conflicts from previous ingestion (there is no need to reuse the cache from a previous ingestion)
	log.I.Info("Preparing cache provider")
	err = g.providers.CacheProvider.Prepare(runCtx) //nolint: contextcheck
	if err != nil {
		return fmt.Errorf("cache client creation: %w", err)
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
	alreadyIngestedInDB, err := g.isAlreadyIngestedInDB(runCtx, clusterName, runID) //nolint: contextcheck
	if err != nil {
		return err
	}

	if alreadyIngestedInDB {
		log.I.Infof("Data already ingested in the database for %s/%s, droping the current data", clusterName, runID)
		err := g.providers.StoreProvider.Clean(runCtx, runID, clusterName) //nolint: contextcheck
		if err != nil {
			return err
		}
	}

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

func (g *IngestorAPI) isAlreadyIngestedInGraph(_ context.Context, clusterName string, runID string) (bool, error) {
	var err error
	gClient, ok := g.providers.GraphProvider.Raw().(*gremlingo.DriverRemoteConnection)
	if !ok {
		return false, fmt.Errorf("assert gClient as *gremlingo.DriverRemoteConnection")
	}

	gQuery := gremlingo.Traversal_().WithRemote(gClient)

	// Using nodes as it should be the "smallest" type of asset in the graph
	rawCount, err := gQuery.V().Has("runID", runID).Has("cluster", clusterName).Limit(1).Count().Next()
	if err != nil {
		return false, fmt.Errorf("getting nodes for %s/%s: %w", runID, clusterName, err)
	}
	nodeCount, err := rawCount.GetInt()
	if err != nil {
		return false, fmt.Errorf("counting nodes for %s/%s: %w", runID, clusterName, err)
	}

	if nodeCount != 0 {
		return true, nil
	}

	return false, nil
}

func (g *IngestorAPI) isAlreadyIngestedInDB(ctx context.Context, clusterName string, runID string) (bool, error) {
	var resNum int64
	var err error
	for _, collection := range collections.GetCollections() {
		mdb := adapter.MongoDB(g.providers.StoreProvider)
		db := mdb.Collection(collection)
		filter := bson.M{
			"runtime": bson.M{
				"runID":   runID,
				"cluster": clusterName,
			},
		}
		resNum, err = db.CountDocuments(ctx, filter, nil)
		if err != nil {
			return false, fmt.Errorf("error counting documents in collection %s: %w", collection, err)
		}
		if resNum != 0 {
			log.I.Infof("Found %d element in collection %s", resNum, collection)

			return true, nil
		}
		log.I.Debugf("Found %d element in collection %s", resNum, collection)
	}

	return false, nil
}

// Notify notifies the caller that the ingestion is completed
func (g *IngestorAPI) Notify(ctx context.Context, clusterName string, runID string) error {
	return g.notifier.Notify(ctx, clusterName, runID)
}
