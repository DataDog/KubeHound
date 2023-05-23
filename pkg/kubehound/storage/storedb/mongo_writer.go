package storedb

import (
	"context"
	"fmt"
	"math"

	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var _ AsyncWriter = (*MongoAsyncWriter)(nil)

type MongoAsyncWriter struct {
	mongodb    *MongoProvider
	ops        []mongo.WriteModel
	collection *mongo.Collection
	batchSize  int
}

// Queue add a model to an asynchronous write queue. Non-blocking.
func (maw *MongoAsyncWriter) Queue(ctx context.Context, model any) error {
	maw.ops = append(maw.ops, mongo.NewInsertOneModel().SetDocument(model))
	return nil
}

// Flush triggers writes of any remaining items in the queue.
// wait on the returned channel which will be signaled when the flush operation completes.
func (maw *MongoAsyncWriter) Flush(ctx context.Context) (chan struct{}, error) {
	ch := make(chan struct{}, 0)
	defer func(chan struct{}) {
		log.I.Infof("Flushed data from mongo writer")
		ch <- struct{}{}
	}(ch)

	if maw.mongodb.client == nil {
		return ch, fmt.Errorf("mongodb client is not initialized")
	}

	if maw.collection == nil {
		return ch, fmt.Errorf("mongodb collection is not initialized")
	}

	if len(maw.ops) == 0 {
		log.I.Debugf("Skipping flush on %s as no write operations", maw.collection.Name())
		return ch, nil
	}

	for start := 0; start < len(maw.ops); start += maw.batchSize {
		end := start + maw.batchSize
		// Avoid overflow
		end = int(math.Min(float64(end), float64(len(maw.ops))))
		// Nothing left in the queue, don't send it through mongodb driver!
		if start == end {
			break
		}

		bulkWriteOpts := options.BulkWrite().SetOrdered(false)
		_, err := maw.collection.BulkWrite(ctx, maw.ops[start:end], bulkWriteOpts)
		if err != nil {
			log.I.Errorf("could not write in bulk to mongo: %w", err)
		}
	}
	return ch, nil
}

// Close cleans up any resources used by the AsyncWriter implementation. Writer cannot be reused after this call.
func (maw *MongoAsyncWriter) Close(ctx context.Context) error {
	if maw.mongodb.client == nil {
		return nil
	}

	err := maw.mongodb.client.Disconnect(ctx)
	if err != nil {
		return err
	}

	maw.ops = nil
	return nil
}
