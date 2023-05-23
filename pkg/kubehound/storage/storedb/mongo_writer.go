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
// On error, we return a nil channel, otherwise we always send something to the channel
func (maw *MongoAsyncWriter) Flush(ctx context.Context) (chan struct{}, error) {
	ch := make(chan struct{}, 1)

	if maw.mongodb.client == nil {
		return nil, fmt.Errorf("mongodb client is not initialized")
	}

	if maw.collection == nil {
		return nil, fmt.Errorf("mongodb collection is not initialized")
	}

	if len(maw.ops) == 0 {
		log.I.Debugf("Skipping flush on %s as no write operations", maw.collection.Name())
		// we need to send something to the channel from this function whenever we don't return an error
		// we cannot defer it because the go routine may last longer than the current function
		// the defer is going to be executed at the return time, whetever or not the inner go routine is processing data
		ch <- struct{}{}
		return ch, nil
	}

	go func(chan struct{}) {
		for start := 0; start < len(maw.ops); start += maw.batchSize {
			// time.Sleep(5 * time.Second)
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
		// we only clear the ops slice after we have finished all bulk writes
		maw.ops = nil
		ch <- struct{}{}
	}(ch)
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
