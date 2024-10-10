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
	l := log.Logger(context.TODO())
	db, ok := store.Reader().(*mongo.Database)
	if !ok {
		l.Fatalf("Invalid database provider type. Expected *mongo.Client, got %T", store.Reader())
	}

	return db
}

// MongoCursorHandler is the default stream implementation to handle the query results from a mongoDB store provider.
func MongoCursorHandler[T any](ctx context.Context, cur *mongo.Cursor,
	callback types.ProcessEntryCallback, complete types.CompleteQueryCallback) error {

	for cur.Next(ctx) {
		var entry T
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
