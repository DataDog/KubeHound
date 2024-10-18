package storedb

import (
	"context"
	"fmt"
	"sync"

	"github.com/DataDog/KubeHound/pkg/kubehound/store/collections"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	"github.com/DataDog/KubeHound/pkg/telemetry/metric"
	"github.com/DataDog/KubeHound/pkg/telemetry/span"
	"github.com/DataDog/KubeHound/pkg/telemetry/statsd"
	"github.com/DataDog/KubeHound/pkg/telemetry/tag"
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
	collection      collections.Collection
	ops             []mongo.WriteModel
	opsLock         *sync.RWMutex
	dbWriter        *mongo.Collection
	batchSize       int
	consumerChan    chan []mongo.WriteModel
	writingInFlight *sync.WaitGroup
	tags            []string
}

func NewMongoAsyncWriter(ctx context.Context, db *mongo.Database, collection collections.Collection, opts ...WriterOption) *MongoAsyncWriter {
	wOpts := &writerOptions{}
	for _, o := range opts {
		o(wOpts)
	}

	maw := MongoAsyncWriter{
		dbWriter:        db.Collection(collection.Name()),
		batchSize:       collection.BatchSize(),
		tags:            append(wOpts.Tags, tag.Collection(collection.Name())),
		collection:      collection,
		writingInFlight: &sync.WaitGroup{},
		ops:             make([]mongo.WriteModel, 0),
		opsLock:         &sync.RWMutex{},
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

				_ = statsd.Count(metric.BackgroundWriterCall, 1, maw.tags, 1)
				err := maw.batchWrite(ctx, data)
				if err != nil {
					log.Trace(ctx).Errorf("write data in background batch writer: %v", err)
				}

				_ = statsd.Decr(metric.QueueSize, maw.tags, 1)
			case <-ctx.Done():
				log.Trace(ctx).Debug("Closed background mongodb worker")

				return
			}
		}
	}()
}

// batchWrite blocks until the write is complete
func (maw *MongoAsyncWriter) batchWrite(ctx context.Context, ops []mongo.WriteModel) error {
	span, ctx := span.SpanRunFromContext(ctx, span.MongoDBBatchWrite)
	span.SetTag(tag.CollectionTag, maw.collection.Name())
	var err error
	defer func() { span.Finish(tracer.WithError(err)) }()
	defer maw.writingInFlight.Done()

	_ = statsd.Count(metric.ObjectWrite, int64(len(ops)), maw.tags, 1)

	bulkWriteOpts := options.BulkWrite().SetOrdered(false)
	_, err = maw.dbWriter.BulkWrite(ctx, ops, bulkWriteOpts)
	if err != nil {
		return fmt.Errorf("could not write in bulk to mongo: %w", err)
	}

	return nil
}

// Queue add a model to an asynchronous write queue. Non-blocking.
func (maw *MongoAsyncWriter) Queue(ctx context.Context, model any) error {
	maw.opsLock.Lock()
	defer maw.opsLock.Unlock()

	maw.ops = append(maw.ops, mongo.NewInsertOneModel().SetDocument(model))
	if len(maw.ops) > maw.batchSize {
		copied := make([]mongo.WriteModel, len(maw.ops))
		copy(copied, maw.ops)

		maw.writingInFlight.Add(1)
		maw.consumerChan <- copied
		_ = statsd.Incr(metric.QueueSize, maw.tags, 1)

		// cleanup the ops array after we have copied it to the channel
		maw.ops = nil
	}

	return nil
}

// Flush triggers writes of any remaining items in the queue.
// This is blocking
func (maw *MongoAsyncWriter) Flush(ctx context.Context) error {
	span, ctx := span.SpanRunFromContext(ctx, span.MongoDBFlush)
	span.SetTag(tag.CollectionTag, maw.collection.Name())
	var err error
	defer func() { span.Finish(tracer.WithError(err)) }()

	if maw.dbWriter == nil {
		return fmt.Errorf("mongodb client is not initialized")
	}

	if maw.collection == nil {
		return fmt.Errorf("mongodb collection is not initialized")
	}

	maw.opsLock.Lock()
	defer maw.opsLock.Unlock()

	if len(maw.ops) != 0 {
		maw.writingInFlight.Add(1)
		err = maw.batchWrite(ctx, maw.ops)
		if err != nil {
			log.Trace(ctx).Errorf("batch write %s: %+v", maw.collection.Name(), err)
			maw.writingInFlight.Wait()

			return err
		}

		maw.ops = nil
	}

	maw.writingInFlight.Wait()

	return nil
}

// Close cleans up any resources used by the AsyncWriter implementation. Writer cannot be reused after this call.
func (maw *MongoAsyncWriter) Close(ctx context.Context) error {
	if maw.dbWriter == nil {
		return nil
	}

	maw.ops = nil

	return nil
}
