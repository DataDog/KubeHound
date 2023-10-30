package storedb

import (
	"context"
	"fmt"
	"time"

	"github.com/DataDog/KubeHound/pkg/kubehound/store/collections"
	"github.com/DataDog/KubeHound/pkg/telemetry/tag"
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

type MongoProvider struct {
	// client *mongo.Client
	// db     *mongo.Database
	reader *mongo.Client
	writer *mongo.Client
	tags   []string
}

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

func createReader(ctx context.Context, url string, timeout time.Duration) (*mongo.Client, error) {
	opts := options.Client()
	opts.ApplyURI(url + fmt.Sprintf("/?connectTimeoutMS=%d", timeout))
	opts.Monitor = mongotrace.NewMonitor()

	return createClient(ctx, opts, timeout)
}

func createWriter(ctx context.Context, url string, timeout time.Duration) (*mongo.Client, error) {
	opts := options.Client()
	opts.ApplyURI(url + fmt.Sprintf("/?connectTimeoutMS=%d", timeout))

	return createClient(ctx, opts, timeout)
}

func NewMongoProvider(ctx context.Context, url string, connectionTimeout time.Duration) (*MongoProvider, error) {
	opts := options.Client()
	opts.ApplyURI(url + fmt.Sprintf("/?connectTimeoutMS=%d", connectionTimeout))
	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, connectionTimeout)
	defer cancel()
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		return nil, err
	}

	db := client.Database(MongoDatabaseName)

	return &MongoProvider{
		client: client,
		db:     db,
		tags:   append(tag.BaseTags, tag.Storage(StorageProviderName)),
	}, nil
}

func (mp *MongoProvider) Prepare(ctx context.Context) error {
	collections, err := mp.db.ListCollectionNames(ctx, bson.M{})
	if err != nil {
		return fmt.Errorf("listing mongo DB collections: %w", err)
	}

	for _, collectionName := range collections {
		err = mp.db.Collection(collectionName).Drop(ctx)
		if err != nil {
			return fmt.Errorf("deleting mongo DB collection %s: %w", collectionName, err)
		}
	}

	ib, err := NewIndexBuilder(mp.db)
	if err != nil {
		return fmt.Errorf("mongo DB index builder create: %w", err)
	}

	if err := ib.BuildAll(ctx); err != nil {
		return fmt.Errorf("mongo DB index builder run: %w", err)
	}

	return nil
}

func (mp *MongoProvider) Raw() any {
	return mp.client
}

func (mp *MongoProvider) Name() string {
	return StorageProviderName
}

func (mp *MongoProvider) HealthCheck(ctx context.Context) (bool, error) {
	err := mp.client.Ping(ctx, nil)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (mp *MongoProvider) Close(ctx context.Context) error {
	errors := errors.Mu
	if mp.reader != nil {
		err := mp.reader.Disconnect(ctx)
	}
	err := mp.reader.Disconnect(ctx)
	return mp.client.Disconnect(ctx)
}

func (mp *MongoProvider) BulkWriter(ctx context.Context, collection collections.Collection, opts ...WriterOption) (AsyncWriter, error) {
	writer := NewMongoAsyncWriter(ctx, mp, collection, opts...)

	return writer, nil
}
