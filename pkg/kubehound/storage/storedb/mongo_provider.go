package storedb

import (
	"context"
	"fmt"
	"time"

	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/DataDog/KubeHound/pkg/kubehound/store/collections"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	"github.com/DataDog/KubeHound/pkg/telemetry/tag"
	"github.com/hashicorp/go-multierror"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	mongotrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/go.mongodb.org/mongo-driver/mongo"
)

const (
	StorageProviderName = "mongodb"
)

var (
	_ Provider = (*MongoProvider)(nil)
)

// A MongoDB based store provider implementation.
type MongoProvider struct {
	reader *mongo.Client // MongoDB client optimized for read operations
	writer *mongo.Client // MongoDB client optimized for write operations
	tags   []string      // Tags to be applied for telemetry
}

// createClient creates a new MongoDB client with the provided options.
func createClient(ctx context.Context, opts *options.ClientOptions, timeout time.Duration) (*mongo.Client, error) {
	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		return nil, err
	}

	return client, nil
}

// createReaderWriter creates a pair of MongoDB clients - one for writes and another for reads.
func createReaderWriter(ctx context.Context, url string, timeout time.Duration) (*mongo.Client, *mongo.Client, error) {
	baseOpts := options.Client()
	baseOpts.ApplyURI(url + fmt.Sprintf("/?connectTimeoutMS=%d", timeout))

	writer, err := createClient(ctx, baseOpts, timeout)
	if err != nil {
		return nil, nil, err
	}

	opts := baseOpts
	opts.Monitor = mongotrace.NewMonitor()
	reader, err := createClient(ctx, opts, timeout)
	if err != nil {
		_ = writer.Disconnect(ctx)

		return nil, nil, err
	}

	return reader, writer, nil
}

// NewMongoProvider creates a new instance of the MongoDB store provider
func NewMongoProvider(ctx context.Context, cfg *config.KubehoundConfig) (*MongoProvider, error) {
	reader, writer, err := createReaderWriter(ctx, cfg.MongoDB.URL, cfg.MongoDB.ConnectionTimeout)
	if err != nil {
		return nil, err
	}

	return &MongoProvider{
		reader: reader,
		writer: writer,
		tags:   tag.GetBaseTagsWith(tag.Storage(StorageProviderName)),
	}, nil
}

func (mp *MongoProvider) Prepare(ctx context.Context) error {
	db := mp.writer.Database(MongoDatabaseName)
	collections, err := db.ListCollectionNames(ctx, bson.M{})
	if err != nil {
		return fmt.Errorf("listing mongo DB collections: %w", err)
	}

	for _, collectionName := range collections {
		err = db.Collection(collectionName).Drop(ctx)
		if err != nil {
			return fmt.Errorf("deleting mongo DB collection %s: %w", collectionName, err)
		}
	}

	ib, err := NewIndexBuilder(db)
	if err != nil {
		return fmt.Errorf("mongo DB index builder create: %w", err)
	}

	if err := ib.BuildAll(ctx); err != nil {
		return fmt.Errorf("mongo DB index builder run: %w", err)
	}

	return nil
}

func (mp *MongoProvider) Clean(ctx context.Context, runId string, cluster string) error {
	db := mp.writer.Database(MongoDatabaseName)
	collections, err := db.ListCollectionNames(ctx, bson.M{})
	if err != nil {
		return fmt.Errorf("listing mongo DB collections: %w", err)
	}
	filter := bson.M{
		"runtime.runID":   runId,
		"runtime.cluster": cluster,
	}
	for _, collectionName := range collections {
		res, err := db.Collection(collectionName).DeleteMany(ctx, filter)
		if err != nil {
			return fmt.Errorf("deleting mongo DB collection %s: %w", collectionName, err)
		}
		log.I.Infof("Deleted %d elements from collection %s", res.DeletedCount, collectionName)
	}

	return nil
}

func (mp *MongoProvider) Reader() any {
	return mp.reader.Database(MongoDatabaseName)
}

func (mp *MongoProvider) Name() string {
	return StorageProviderName
}

func (mp *MongoProvider) HealthCheck(ctx context.Context) (bool, error) {
	err := mp.reader.Ping(ctx, nil)
	if err != nil {
		return false, err
	}

	err = mp.writer.Ping(ctx, nil)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (mp *MongoProvider) Close(ctx context.Context) error {
	var res *multierror.Error
	if mp.reader != nil {
		err := mp.reader.Disconnect(ctx)
		if err != nil {
			res = multierror.Append(res, err)
		}
	}

	if mp.writer != nil {
		err := mp.writer.Disconnect(ctx)
		if err != nil {
			res = multierror.Append(res, err)
		}
	}

	return res.ErrorOrNil()
}

func (mp *MongoProvider) BulkWriter(ctx context.Context, collection collections.Collection, opts ...WriterOption) (AsyncWriter, error) {
	writer := NewMongoAsyncWriter(ctx, mp.writer.Database(MongoDatabaseName), collection, opts...)

	return writer, nil
}
