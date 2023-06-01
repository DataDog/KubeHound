package storedb

import (
	"context"
	"fmt"
	"time"

	"github.com/DataDog/KubeHound/pkg/kubehound/store/collections"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	mongotrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/go.mongodb.org/mongo-driver/mongo"
)

var _ Provider = (*MongoProvider)(nil)

type MongoProvider struct {
	client *mongo.Client
	db     *mongo.Database
}

func NewMongoProvider(ctx context.Context, url string, connectionTimeout time.Duration) (*MongoProvider, error) {
	opts := options.Client()
	opts.Monitor = mongotrace.NewMonitor()
	opts.ApplyURI(url + fmt.Sprintf("/?connectTimeoutMS=%d", connectionTimeout))

	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		return nil, err
	}

	db := client.Database(MongoDatabaseName)
	return &MongoProvider{
		client: client,
		db:     db,
	}, nil
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
	writer := NewMongoAsyncWriter(ctx, mp, collection)
	return writer, nil
}
