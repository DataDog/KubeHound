package storedb

import (
	"context"
	"fmt"
	"sync"

	"github.com/DataDog/KubeHound/pkg/kubehound/store/collections"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var _ AsyncWriter = (*MongoAsyncWriter)(nil)

type MongoAsyncWriter struct {
	mongodb         *MongoProvider
	ops             []mongo.WriteModel
	collection      *mongo.Collection
	batchSize       int
	consummerChan   chan []mongo.WriteModel
	writingInFligth sync.WaitGroup
}

func NewMongoAsyncWriter(ctx context.Context, mp *MongoProvider, collection collections.Collection) *MongoAsyncWriter {
	maw := MongoAsyncWriter{
		mongodb:    mp,
		collection: mp.db.Collection(collection.Name()),
		batchSize:  collection.BatchSize(),
	}
	// creating a buffered channel of size one.
	maw.consummerChan = make(chan []mongo.WriteModel, 1)
	maw.backgroundWriter(ctx)
	return &maw
}

// backgroundWriter starts a background go routine
func (maw *MongoAsyncWriter) backgroundWriter(ctx context.Context) {
	go func() {
		for {
			select {
			case data := <-maw.consummerChan:
				// closing the channel shoud stop the go routine
				if data == nil {
					return
				}
				err := maw.batchWrite(ctx, data)
				if err != nil {
					log.I.Errorf("failed to write data in background batch writer: %w", err)
				}
			case <-ctx.Done():
				log.I.Info("Closed background mongodb worker")
				return
			}
		}
	}()
}

// batchWrite blocks until the write is complete
func (maw *MongoAsyncWriter) batchWrite(ctx context.Context, ops []mongo.WriteModel) error {
	maw.writingInFligth.Add(1)
	defer maw.writingInFligth.Done()
	bulkWriteOpts := options.BulkWrite().SetOrdered(false)
	_, err := maw.collection.BulkWrite(ctx, ops, bulkWriteOpts)
	if err != nil {
		return fmt.Errorf("could not write in bulk to mongo: %w", err)
	}
	return nil
}

// Queue add a model to an asynchronous write queue. Non-blocking.
func (maw *MongoAsyncWriter) Queue(ctx context.Context, model any) error {
	if len(maw.ops) > maw.batchSize {
		maw.consummerChan <- maw.ops
		// cleanup the ops array after we have copied it to the channel
		maw.ops = nil
	}
	maw.ops = append(maw.ops, mongo.NewInsertOneModel().SetDocument(model))
	return nil
}

// Flush triggers writes of any remaining items in the queue.
// This is blocking
func (maw *MongoAsyncWriter) Flush(ctx context.Context) error {
	if maw.mongodb.client == nil {
		return fmt.Errorf("mongodb client is not initialized")
	}

	if maw.collection == nil {
		return fmt.Errorf("mongodb collection is not initialized")
	}

	if len(maw.ops) == 0 {
		log.I.Debugf("Skipping flush on %s as no write operations", maw.collection.Name())
		// we need to send something to the channel from this function whenever we don't return an error
		// we cannot defer it because the go routine may last longer than the current function
		// the defer is going to be executed at the return time, whetever or not the inner go routine is processing data
		maw.writingInFligth.Wait()
		return nil
	}

	err := maw.batchWrite(ctx, maw.ops)
	if err != nil {
		maw.writingInFligth.Wait()
		return err
	}

	maw.ops = nil
	return nil
}

// Close cleans up any resources used by the AsyncWriter implementation. Writer cannot be reused after this call.
func (maw *MongoAsyncWriter) Close(ctx context.Context) error {
	if maw.mongodb.client == nil {
		return nil
	}

	maw.ops = nil
	return nil
}
