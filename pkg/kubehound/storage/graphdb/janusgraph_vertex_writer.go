package graphdb

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/DataDog/KubeHound/pkg/kubehound/graph/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/vertex"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/cache"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/cache/cachekey"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	"github.com/DataDog/KubeHound/pkg/telemetry/metric"
	"github.com/DataDog/KubeHound/pkg/telemetry/span"
	"github.com/DataDog/KubeHound/pkg/telemetry/statsd"
	"github.com/DataDog/KubeHound/pkg/telemetry/tag"
	gremlin "github.com/apache/tinkerpop/gremlin-go/v3/driver"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

var _ AsyncVertexWriter = (*JanusGraphVertexWriter)(nil)

type JanusGraphVertexWriter struct {
	builder         string                          // Name of the graph entity being written
	gremlin         types.VertexTraversal           // Gremlin traversal generator function
	drc             *gremlin.DriverRemoteConnection // Gremlin driver remote connection
	traversalSource *gremlin.GraphTraversalSource   // Transacted graph traversal source
	writingInFlight *sync.WaitGroup                 // Wait group tracking current unfinished writes
	qcounter        int32                           // Track items queued
	wcounter        int32                           // Track items writtn
	tags            []string                        // Telemetry tags
	cache           cache.AsyncWriter               // Cache writer to cache store id -> vertex id mappings
	writerTimeout   time.Duration                   // Timeout for the writer
	maxRetry        int                             // Maximum number of retries for failed writes
	mb              *microBatcher                   // Micro batcher to batch writes
}

// NewJanusGraphAsyncVertexWriter creates a new bulk vertex writer instance.
func NewJanusGraphAsyncVertexWriter(ctx context.Context, drc *gremlin.DriverRemoteConnection,
	v vertex.Builder, c cache.CacheProvider, opts ...WriterOption,
) (*JanusGraphVertexWriter, error) {
	options := &writerOptions{
		WriterTimeout:     defaultWriterTimeout,
		MaxRetry:          defaultMaxRetry,
		WriterWorkerCount: defaultWriterWorkerCount,
	}
	for _, opt := range opts {
		opt(options)
	}

	cw, err := c.BulkWriter(ctx, cache.WithTest())
	if err != nil {
		return nil, fmt.Errorf("janusgraph vertex writer cache creation: %w", err)
	}

	jw := JanusGraphVertexWriter{
		builder:         v.Label(),
		gremlin:         v.Traversal(),
		drc:             drc,
		traversalSource: gremlin.Traversal_().WithRemote(drc),
		writingInFlight: &sync.WaitGroup{},
		tags:            append(options.Tags, tag.Label(v.Label()), tag.Builder(v.Label())),
		cache:           cw,
		writerTimeout:   options.WriterTimeout,
		maxRetry:        options.MaxRetry,
	}

	// Create a new micro batcher to batch the inserts with split and retry logic.
	jw.mb = newMicroBatcher(log.Trace(ctx), v.BatchSize(), options.WriterWorkerCount, func(ctx context.Context, a []any) error {
		// Increment the writingInFlight wait group to track the number of writes in progress.
		jw.writingInFlight.Add(1)
		defer jw.writingInFlight.Done()

		// Try to write the batch to the graph DB.
		if err := jw.batchWrite(ctx, a); err != nil {
			var bwe *batchWriterError
			if errors.As(err, &bwe) && bwe.retryable {
				// If the write operation failed and is retryable, split the batch and retry.
				return jw.splitAndRetry(ctx, 0, a)
			}

			return err
		}

		return nil
	})
	jw.mb.Start(ctx)

	return &jw, nil
}

func (jgv *JanusGraphVertexWriter) cacheIds(ctx context.Context, idMap []*gremlin.Result) error {
	for _, r := range idMap {
		idMap, ok := r.GetInterface().(map[interface{}]interface{})
		if !ok {
			return fmt.Errorf("parsing vertex insert result map: %#v", r)
		}

		storeID, sOk := idMap["storeID"].(string)
		vertexId, vOk := idMap["id"].(int64)

		if !sOk || !vOk {
			return errors.New("vertex id type conversion")
		}

		err := jgv.cache.Queue(ctx, cachekey.ObjectID(storeID), vertexId)
		if err != nil {
			return fmt.Errorf("vertex id cache write: %w", err)
		}
	}

	return nil
}

// batchWrite will write a batch of entries into the graph DB and block until the write completes.
func (jgv *JanusGraphVertexWriter) batchWrite(ctx context.Context, data []any) error {
	_ = statsd.Count(ctx, metric.BackgroundWriterCall, 1, jgv.tags, 1)

	span, ctx := span.SpanRunFromContext(ctx, span.JanusGraphBatchWrite)
	span.SetTag(tag.LabelTag, jgv.builder)
	var err error
	defer func() { span.Finish(tracer.WithError(err)) }()

	datalen := len(data)
	_ = statsd.Count(ctx, metric.VertexWrite, int64(datalen), jgv.tags, 1)
	log.Trace(ctx).Debugf("Batch write JanusGraphVertexWriter with %d elements", datalen)
	atomic.AddInt32(&jgv.wcounter, int32(datalen)) //nolint:gosec // disable G115

	// Create a channel to signal the completion of the write operation.
	errChan := make(chan error, 1)

	// We need to ensure that the write operation is completed within a certain
	// time frame to avoid blocking the writer indefinitely if the backend
	// is unresponsive.
	go func() {
		// Create a new gremlin operation to insert the data into the graph.
		op := jgv.gremlin(jgv.traversalSource, data)
		raw, err := op.Project("id", "storeID").
			By(gremlin.T.Id).
			By("storeID").
			ToList()
		if err != nil {
			errChan <- fmt.Errorf("%s vertex insert: %w", jgv.builder, err)

			return
		}

		// Gremlin will return a list of maps containing and vertex id and store
		// id values for each vertex inserted.
		// We need to parse each map entry and add to our cache.
		if err = jgv.cacheIds(ctx, raw); err != nil {
			errChan <- fmt.Errorf("cache ids: %w", err)

			return
		}

		errChan <- nil
	}()

	// Wait for the write operation to complete or timeout.
	select {
	case <-ctx.Done():
		// If the context is cancelled, return the error.
		return ctx.Err()
	case <-time.After(jgv.writerTimeout):
		// If the write operation takes too long, return an error.
		return &batchWriterError{
			err:       errors.New("vertex write operation timed out"),
			retryable: true,
		}
	case err = <-errChan:
		if err != nil {
			return fmt.Errorf("janusgraph batch write: %w", err)
		}
	}

	return nil
}

// retrySplitAndRequeue will split the batch into smaller chunks and requeue them for writing.
func (jgv *JanusGraphVertexWriter) splitAndRetry(ctx context.Context, retryCount int, payload []any) error {
	_ = statsd.Count(ctx, metric.RetryWriterCall, 1, jgv.tags, 1)

	// If we have reached the maximum number of retries, return an error.
	if retryCount >= jgv.maxRetry {
		return fmt.Errorf("max retry count reached: %d", retryCount)
	}

	// Compute the new batch size.
	newBatchSize := len(payload) / 2

	log.Trace(ctx).Warnf("Retrying write operation with smaller vertex batch (n:%d -> %d, r:%d)", len(payload), newBatchSize, retryCount)

	var leftErr, rightErr error

	// Split the batch into smaller chunks and retry them.
	if len(payload[:newBatchSize]) > 0 {
		if leftErr = jgv.batchWrite(ctx, payload[:newBatchSize]); leftErr == nil {
			var bwe *batchWriterError
			if errors.As(leftErr, &bwe) && bwe.retryable {
				return jgv.splitAndRetry(ctx, retryCount+1, payload[:newBatchSize])
			}
		}
	}

	// Process the right side of the batch.
	if len(payload[newBatchSize:]) > 0 {
		if rightErr = jgv.batchWrite(ctx, payload[newBatchSize:]); rightErr != nil {
			var bwe *batchWriterError
			if errors.As(rightErr, &bwe) && bwe.retryable {
				return jgv.splitAndRetry(ctx, retryCount+1, payload[newBatchSize:])
			}
		}
	}

	// Return the first error encountered.
	switch {
	case leftErr != nil && rightErr != nil:
		return fmt.Errorf("left: %w, right: %w", leftErr, rightErr)
	case leftErr != nil:
		return leftErr
	case rightErr != nil:
		return rightErr
	}

	return nil
}

func (jgv *JanusGraphVertexWriter) Close(ctx context.Context) error {
	if jgv.cache != nil {
		if err := jgv.cache.Close(ctx); err != nil {
			return fmt.Errorf("closing cache: %w", err)
		}
	}

	return nil
}

// Flush triggers writes of any remaining items in the queue.
// This is blocking
func (jgv *JanusGraphVertexWriter) Flush(ctx context.Context) error {
	span, ctx := span.SpanRunFromContext(ctx, span.JanusGraphFlush)
	span.SetTag(tag.LabelTag, jgv.builder)
	var err error
	defer func() { span.Finish(tracer.WithError(err)) }()

	if jgv.traversalSource == nil {
		return errors.New("janusGraph traversalSource is not initialized")
	}

	// Flush the micro batcher.
	err = jgv.mb.Flush(ctx)
	if err != nil {
		return fmt.Errorf("micro batcher flush: %w", err)
	}

	// Wait for all writes to complete.
	jgv.writingInFlight.Wait()

	err = jgv.cache.Flush(ctx)
	if err != nil {
		return fmt.Errorf("vertex id cacheflush: %w", err)
	}

	log.Trace(ctx).Debugf("Batch writer %d %s queued", jgv.qcounter, jgv.builder)
	log.Trace(ctx).Infof("Batch writer %d %s written", jgv.wcounter, jgv.builder)

	return nil
}

func (jgv *JanusGraphVertexWriter) Queue(ctx context.Context, v any) error {
	atomic.AddInt32(&jgv.qcounter, 1)

	return jgv.mb.Enqueue(ctx, v)
}
