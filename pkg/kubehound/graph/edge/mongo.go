package edge

import (
	"context"
	"fmt"

	"github.com/DataDog/KubeHound/pkg/kubehound/storage/storedb"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	"github.com/DataDog/KubeHound/pkg/utils"
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

// MongoProcessor is the default processor implementation for a mongoDB store provider.
func MongoProcessor[T any](_ context.Context, entry DataContainer) (map[string]any, error) {
	typed, ok := entry.(T)
	if !ok {
		return nil, fmt.Errorf("invalid type passed to processor: %T", entry)
	}

	processed, err := utils.StructToMap(typed)
	if err != nil {
		return nil, err
	}

	return processed, nil
}

// MongoCursorHandler is the default stream implementation to handle the query results from a mongoDB store provider.
func MongoCursorHandler[T any](ctx context.Context, cur *mongo.Cursor,
	callback ProcessEntryCallback, complete CompleteQueryCallback) error {

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
