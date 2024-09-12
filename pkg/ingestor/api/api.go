package api

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"github.com/DataDog/KubeHound/pkg/collector"
	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/DataDog/KubeHound/pkg/dump"
	"github.com/DataDog/KubeHound/pkg/ingestor"
	grpc "github.com/DataDog/KubeHound/pkg/ingestor/api/grpc/pb"
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
	"google.golang.org/protobuf/types/known/timestamppb"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

type API interface {
	Ingest(ctx context.Context, path string) error
	Notify(ctx context.Context, clusterName string, runID string) error
}

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

// RehydrateLatest is just a GRPC wrapper around the Ingest method from the API package
func (g *IngestorAPI) RehydrateLatest(ctx context.Context) ([]*grpc.IngestedCluster, error) {
	// first level key are cluster names
	directories, errRet := g.puller.ListFiles(ctx, "", false)
	if errRet != nil {
		return nil, errRet
	}

	res := []*grpc.IngestedCluster{}

	for _, dir := range directories {
		clusterName := strings.TrimSuffix(dir.Key, "/")

		dumpKeys, err := g.puller.ListFiles(ctx, clusterName, true)
		if err != nil {
			return nil, err
		}

		if l := len(dumpKeys); l > 0 {
			// extracting the latest runID
			latestDump := slices.MaxFunc(dumpKeys, func(a, b *puller.ListObject) int {
				// return dumpKeys[a].ModTime.Before(dumpKeys[b].ModTime)
				return a.ModTime.Compare(b.ModTime)
			})
			latestDumpIngestTime := latestDump.ModTime
			latestDumpKey := latestDump.Key

			clusterErr := g.Ingest(ctx, latestDumpKey)
			if clusterErr != nil {
				errRet = errors.Join(errRet, fmt.Errorf("ingesting cluster %s: %w", latestDumpKey, clusterErr))
			}
			log.I.Infof("Rehydrated cluster: %s, date: %s, key: %s", clusterName, latestDumpIngestTime.Format("01-02-2006 15:04:05"), latestDumpKey)
			ingestedCluster := &grpc.IngestedCluster{
				ClusterName: clusterName,
				Key:         latestDumpKey,
				Date:        timestamppb.New(latestDumpIngestTime),
			}
			res = append(res, ingestedCluster)

		}
	}

	return res, errRet
}

func (g *IngestorAPI) Ingest(_ context.Context, path string) error {
	// Settings global variables for the run in the context to propagate them to the spans
	runCtx := context.Background()

	archivePath, err := g.puller.Pull(runCtx, path) //nolint: contextcheck
	if err != nil {
		return err
	}

	defer func() {
		err = errors.Join(err, g.puller.Close(runCtx, archivePath))
	}()

	err = g.puller.Extract(runCtx, archivePath) //nolint: contextcheck
	if err != nil {
		return err
	}

	metadataFilePath := filepath.Join(filepath.Dir(archivePath), collector.MetadataPath)
	md, err := dump.ParseMetadata(runCtx, metadataFilePath) //nolint: contextcheck
	if err != nil {
		log.I.Warnf("no metadata has been parsed (old dump format from v1.4.0 or below do not embed metadata information): %v", err)
		// Backward Compatibility: Extracting the metadata from the path
		dumpMetadata, err := dump.ParsePath(path)
		if err != nil {
			log.I.Warn("parsing path for metadata", err)

			return err
		}
		md = dumpMetadata.Metadata
	}
	clusterName := md.ClusterName
	runID := md.RunID

	err = g.Cfg.ComputeDynamic(config.WithClusterName(clusterName), config.WithRunID(runID))
	if err != nil {
		return err
	}

	runCfg := g.Cfg
	runCfg.Collector = config.CollectorConfig{
		Type: config.CollectorTypeFile,
		File: &config.FileCollectorConfig{
			Directory: filepath.Dir(archivePath),
		},
	}

	events.PushEvent(
		fmt.Sprintf("Ingesting cluster %s with runID %s", clusterName, runID),
		fmt.Sprintf("Ingesting cluster %s with runID %s", clusterName, runID),
		[]string{
			tag.IngestionRunID(runID),
		},
	)

	runCtx = context.WithValue(runCtx, span.ContextLogFieldClusterName, clusterName)
	runCtx = context.WithValue(runCtx, span.ContextLogFieldRunID, runID)

	spanJob, runCtx := span.SpanIngestRunFromContext(runCtx, span.IngestorStartJob)
	defer func() { spanJob.Finish(tracer.WithError(err)) }()

	alreadyIngested, err := g.isAlreadyIngestedInGraph(runCtx, clusterName, runID) //nolint: contextcheck
	if err != nil {
		return err
	}

	if alreadyIngested {
		return fmt.Errorf("%w [%s:%s]", ErrAlreadyIngested, clusterName, runID)
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

	defer func() {
		err = errors.Join(err, collect.Close(runCtx))
	}()
	log.I.Infof("Loaded %s collector client", collect.Name())

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
		return fmt.Errorf("notifying: %w", err)
	}

	// returning err from the defer functions
	return err
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
