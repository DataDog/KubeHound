package storedb

import (
	"context"
	"fmt"
	"sync"

	"github.com/DataDog/KubeHound/pkg/kubehound/store/collections"
	"github.com/DataDog/KubeHound/pkg/telemetry"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	"github.com/DataDog/KubeHound/pkg/telemetry/statsd"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

const (
	// TODO: this might need to be adjusted in the future, potentially per type of collections
	// We don't have the data yet, so lets just hardcode a relatively high value for now
	consumerChanSize = 10000
)

var _ AsyncWriter = (*MongoAsyncWriter)(nil)

type MongoAsyncWriter struct {
	mongodb         *MongoProvider
	ops             []mongo.WriteModel
	collection      *mongo.Collection
	batchSize       int
	consumerChan    chan []mongo.WriteModel
	writingInFlight sync.WaitGroup
	tags            []string
}

func NewMongoAsyncWriter(ctx context.Context, mp *MongoProvider, collection collections.Collection, opts ...WriterOption) *MongoAsyncWriter {
	wOpts := &writerOptions{}
	for _, o := range opts {
		o(wOpts)
	}

	maw := MongoAsyncWriter{
		mongodb:    mp,
		collection: mp.db.Collection(collection.Name()),
		batchSize:  collection.BatchSize(),
		tags:       wOpts.Tags,
	}
	maw.consumerChan = make(chan []mongo.WriteModel, consumerChanSize)
	maw.startBackgroundWriter(ctx)
	return &maw
}

// startBackgroundWriter starts a background go routine
func (maw *MongoAsyncWriter) startBackgroundWriter(ctx context.Context) {
	go func() {
		for {
			select {
			case data := <-maw.consumerChan:
				// closing the channel shoud stop the go routine
				if data == nil {
					return
				}
				_ = statsd.Count(telemetry.MetricStoredbBackgroundWriterCall, 1, maw.tags, 1)
				err := maw.batchWrite(ctx, data)
				if err != nil {
					log.I.Errorf("write data in background batch writer: %v", err)
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
	span, ctx := tracer.StartSpanFromContext(ctx, telemetry.SpanMongodbOperationBatchWrite, tracer.Measured())
	defer span.Finish()

	maw.writingInFlight.Add(1)
	defer maw.writingInFlight.Done()

	_ = statsd.Gauge(telemetry.MetricStoredbBatchWrite, float64(len(ops)), maw.mongodb.tags, 1)

	bulkWriteOpts := options.BulkWrite().SetOrdered(false)
	_, err := maw.collection.BulkWrite(ctx, ops, bulkWriteOpts)
	if err != nil {
		return fmt.Errorf("could not write in bulk to mongo: %w", err)
	}
	return nil
}

// Queue add a model to an asynchronous write queue. Non-blocking.
func (maw *MongoAsyncWriter) Queue(ctx context.Context, model any) error {
	maw.ops = append(maw.ops, mongo.NewInsertOneModel().SetDocument(model))

	_ = statsd.Gauge(telemetry.MetricStoredbQueueSize, float64(len(maw.ops)), maw.mongodb.tags, 1)

	if len(maw.ops) > maw.batchSize {
		maw.consumerChan <- maw.ops
		// cleanup the ops array after we have copied it to the channel
		maw.ops = nil
	}
	return nil
}

// Flush triggers writes of any remaining items in the queue.
// This is blocking
func (maw *MongoAsyncWriter) Flush(ctx context.Context) error {
	span, ctx := tracer.StartSpanFromContext(ctx, telemetry.SpanMongodbOperationFlush, tracer.Measured())
	defer span.Finish()

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
		maw.writingInFlight.Wait()
		return nil
	}

	err := maw.batchWrite(ctx, maw.ops)
	if err != nil {
		maw.writingInFlight.Wait()
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
