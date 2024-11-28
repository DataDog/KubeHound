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
	inserts         []any                           // Object data to be inserted in the graph
	mu              sync.Mutex                      // Mutex protecting access to the inserts array
	consumerChan    chan batchItem                  // Channel consuming inserts for async writing
	writingInFlight *sync.WaitGroup                 // Wait group tracking current unfinished writes
	batchSize       int                             // Batchsize of graph DB inserts
	qcounter        int32                           // Track items queued
	wcounter        int32                           // Track items writtn
	tags            []string                        // Telemetry tags
	cache           cache.AsyncWriter               // Cache writer to cache store id -> vertex id mappings
	writerTimeout   time.Duration                   // Timeout for the writer
	maxRetry        int                             // Maximum number of retries for failed writes
}

// batchItem is a single item in the batch writer queue that contains the data
// to be written and the number of retries.
type batchItem struct {
	data       []any
	retryCount int
}

// NewJanusGraphAsyncVertexWriter creates a new bulk vertex writer instance.
func NewJanusGraphAsyncVertexWriter(ctx context.Context, drc *gremlin.DriverRemoteConnection,
	v vertex.Builder, c cache.CacheProvider, opts ...WriterOption,
) (*JanusGraphVertexWriter, error) {
	options := &writerOptions{
		WriterTimeout: defaultWriterTimeout,
		MaxRetry:      defaultMaxRetry,
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
		inserts:         make([]any, 0, v.BatchSize()),
		traversalSource: gremlin.Traversal_().WithRemote(drc),
		batchSize:       v.BatchSize(),
		writingInFlight: &sync.WaitGroup{},
		consumerChan:    make(chan batchItem, v.BatchSize()*channelSizeBatchFactor),
		tags:            append(options.Tags, tag.Label(v.Label()), tag.Builder(v.Label())),
		cache:           cw,
		writerTimeout:   options.WriterTimeout,
		maxRetry:        options.MaxRetry,
	}

	jw.startBackgroundWriter(ctx)

	return &jw, nil
}

// startBackgroundWriter starts a background go routine
func (jgv *JanusGraphVertexWriter) startBackgroundWriter(ctx context.Context) {
	go func() {
		for {
			select {
			case batch, ok := <-jgv.consumerChan:
				// If the channel is closed, return.
				if !ok {
					log.Trace(ctx).Info("Closed background janusgraph worker on channel close")
					return
				}

				// If the batch is empty, return.
				if len(batch.data) == 0 {
					log.Trace(ctx).Warn("Empty vertex batch received in background janusgraph worker, skipping")
					return
				}

				_ = statsd.Count(ctx, metric.BackgroundWriterCall, 1, jgv.tags, 1)
				err := jgv.batchWrite(ctx, batch.data)
				if err != nil {
					var e *errBatchWriter
					if errors.As(err, &e) {
						// If the error is retryable, retry the write operation with a smaller batch.
						if e.retryable && batch.retryCount < jgv.maxRetry {
							jgv.retrySplitAndRequeue(ctx, &batch, e)
							continue
						}

						log.Trace(ctx).Errorf("Retry limit exceeded for write operation: %v", err)
					}

					log.Trace(ctx).Errorf("Write data in background batch writer, data will be lost: %v", err)
				}

				_ = statsd.Decr(ctx, metric.QueueSize, jgv.tags, 1)
			case <-ctx.Done():
				log.Trace(ctx).Info("Closed background janusgraph worker on context cancel")

				return
			}
		}
	}()
}

// retrySplitAndRequeue will split the batch into smaller chunks and requeue them for writing.
func (jgv *JanusGraphVertexWriter) retrySplitAndRequeue(ctx context.Context, batch *batchItem, e *errBatchWriter) {
	_ = statsd.Count(ctx, metric.RetryWriterCall, 1, jgv.tags, 1)

	// Compute the new batch size.
	newBatchSize := len(batch.data) / 2
	batch.retryCount++

	log.Trace(ctx).Warnf("Retrying write operation with smaller vertex batch (n:%d -> %d, r:%d): %v", len(batch.data), newBatchSize, batch.retryCount, e.Unwrap())

	// Split the batch into smaller chunks and requeue them.
	if len(batch.data[:newBatchSize]) > 0 {
		jgv.consumerChan <- batchItem{
			data:       batch.data[:newBatchSize],
			retryCount: batch.retryCount,
		}
	}
	if len(batch.data[newBatchSize:]) > 0 {
		jgv.consumerChan <- batchItem{
			data:       batch.data[newBatchSize:],
			retryCount: batch.retryCount,
		}
	}
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
// Callers are responsible for doing an Add(1) to the writingInFlight wait group to ensure proper synchronization.
func (jgv *JanusGraphVertexWriter) batchWrite(ctx context.Context, data []any) error {
	span, ctx := span.SpanRunFromContext(ctx, span.JanusGraphBatchWrite)
	span.SetTag(tag.LabelTag, jgv.builder)
	var err error
	defer func() { span.Finish(tracer.WithError(err)) }()
	defer jgv.writingInFlight.Done()

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
		return &errBatchWriter{
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

func (jgv *JanusGraphVertexWriter) Close(ctx context.Context) error {
	close(jgv.consumerChan)

	return jgv.cache.Close(ctx)
}

// Flush triggers writes of any remaining items in the queue.
// This is blocking
func (jgv *JanusGraphVertexWriter) Flush(ctx context.Context) error {
	span, ctx := span.SpanRunFromContext(ctx, span.JanusGraphFlush)
	span.SetTag(tag.LabelTag, jgv.builder)
	var err error
	defer func() { span.Finish(tracer.WithError(err)) }()

	jgv.mu.Lock()
	defer jgv.mu.Unlock()

	if jgv.traversalSource == nil {
		return errors.New("janusGraph traversalSource is not initialized")
	}

	if len(jgv.inserts) != 0 {
		_ = statsd.Incr(ctx, metric.FlushWriterCall, jgv.tags, 1)

		jgv.writingInFlight.Add(1)
		err = jgv.batchWrite(ctx, jgv.inserts)
		if err != nil {
			log.Trace(ctx).Errorf("batch write %s: %+v", jgv.builder, err)
			jgv.writingInFlight.Wait()

			return err
		}

		log.Trace(ctx).Debugf("Done flushing %s writes. clearing the queue", jgv.builder)
		jgv.inserts = nil
	}

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
	jgv.mu.Lock()
	defer jgv.mu.Unlock()

	atomic.AddInt32(&jgv.qcounter, 1)
	jgv.inserts = append(jgv.inserts, v)

	if len(jgv.inserts) > jgv.batchSize {
		copied := make([]any, len(jgv.inserts))
		copy(copied, jgv.inserts)

		jgv.writingInFlight.Add(1)
		jgv.consumerChan <- batchItem{
			data:       copied,
			retryCount: 0,
		}
		_ = statsd.Incr(ctx, metric.QueueSize, jgv.tags, 1)

		// cleanup the ops array after we have copied it to the channel
		jgv.inserts = nil
	}

	return nil
}
