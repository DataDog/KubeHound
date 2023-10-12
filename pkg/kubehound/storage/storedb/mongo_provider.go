package storedb

import (
	"context"
	"fmt"
	"time"

	"github.com/DataDog/KubeHound/pkg/kubehound/store/collections"
	"github.com/DataDog/KubeHound/pkg/telemetry"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	mongotrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/go.mongodb.org/mongo-driver/mongo"
)

var _ Provider = (*MongoProvider)(nil)

type MongoProvider struct {
	client *mongo.Client
	db     *mongo.Database
	tags   []string
}

func buildIndices(db *mongo.Database, indices collections.IndexSpecRegistry) error {
	mongo.IndexModel
	db.Collection("").Indexes().CreateMany()

	return nil
}

func NewMongoProvider(ctx context.Context, url string, connectionTimeout time.Duration) (*MongoProvider, error) {
	opts := options.Client()
	opts.Monitor = mongotrace.NewMonitor()
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

	// TODO indices creation!

	return &MongoProvider{
		client: client,
		db:     db,
		tags:   []string{telemetry.TagTypeMongodb},
	}, nil
}

func (mp *MongoProvider) Clear(ctx context.Context) error {
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

	return nil
}

func (mp *MongoProvider) Raw() any {
	return mp.client
}

func (mp *MongoProvider) Name() string {
	return "MongoProvider"
}

func (mp *MongoProvider) HealthCheck(ctx context.Context) (bool, error) {
	err := mp.client.Ping(ctx, nil)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (mp *MongoProvider) Close(ctx context.Context) error {
	return mp.client.Disconnect(ctx)
}

func (mp *MongoProvider) BulkWriter(ctx context.Context, collection collections.Collection, opts ...WriterOption) (AsyncWriter, error) {
	writer := NewMongoAsyncWriter(ctx, mp, collection, opts...)

	return writer, nil
}
