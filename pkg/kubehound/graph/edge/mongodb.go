package edge

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/DataDog/KubeHound/pkg/kubehound/storage/storedb"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	"go.mongodb.org/mongo-driver/mongo"
)

func MongoDB(store storedb.Provider) *mongo.Database {
	mongoClient, ok := store.Raw().(*mongo.Client)
	if !ok {
		log.I.Fatalf("Invalid database provider type. Expected *mongo.Client, got %T", store.Raw())
	}

	return mongoClient.Database(storedb.MongoDatabaseName)
}

func MongoProcessor[T any](_ context.Context, entry interface{}) (map[string]any, error) {
	typed, ok := entry.(T)
	if !ok {
		return nil, fmt.Errorf("invalid type passed to processor: %T", entry)
	}

	processed, err := structToMap(typed)
	if err != nil {
		return nil, err
	}

	return processed, nil
}

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

func structToMap(in interface{}) (map[string]any, error) {
	var res map[string]any

	tmp, err := json.Marshal(in)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(tmp, &res)
	if err != nil {
		return nil, err
	}

	return res, nil
}
