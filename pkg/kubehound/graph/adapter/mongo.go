package adapter

import (
	"context"

	"github.com/DataDog/KubeHound/pkg/kubehound/graph/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/storedb"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	"go.mongodb.org/mongo-driver/mongo"
)

// MongoDB is a helper function to retrieve the store database object from a mongoDB provider.
func MongoDB(store storedb.Provider) *mongo.Database {
	mongoClient, ok := store.Raw().(*mongo.Client)
	if !ok {
		log.I.Fatalf("Invalid database provider type. Expected *mongo.Client, got %T", store.Raw())
	}

	return mongoClient.Database(storedb.MongoDatabaseName)
}

// MongoCursorHandler is the default stream implementation to handle the query results from a mongoDB store provider.
func MongoCursorHandler[T any](ctx context.Context, cur *mongo.Cursor,
	callback types.ProcessEntryCallback, complete types.CompleteQueryCallback) error {

	var entry T
	for cur.Next(ctx) {
		err := cur.Decode(&entry)
		if err != nil {
			return err
		}

		err = callback(ctx, &entry)
		if err != nil {
			return err
		}
	}

	return complete(ctx)
}
