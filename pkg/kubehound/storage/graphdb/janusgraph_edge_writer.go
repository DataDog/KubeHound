package graphdb

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/DataDog/KubeHound/pkg/kubehound/graph/edge"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/types"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	"github.com/DataDog/KubeHound/pkg/telemetry/metric"
	"github.com/DataDog/KubeHound/pkg/telemetry/span"
	"github.com/DataDog/KubeHound/pkg/telemetry/statsd"
	"github.com/DataDog/KubeHound/pkg/telemetry/tag"
	gremlingo "github.com/apache/tinkerpop/gremlin-go/v3/driver"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

const (
	TracerServicename = "kubegraph"
)

var _ AsyncEdgeWriter = (*JanusGraphEdgeWriter)(nil)

type JanusGraphEdgeWriter struct {
	builder         string                            // Qualified name of the edge being written
	gremlin         types.EdgeTraversal               // Gremlin traversal generator function
	drc             *gremlingo.DriverRemoteConnection // Gremlin driver remote connection
	traversalSource *gremlingo.GraphTraversalSource   // Transacted graph traversal source
	inserts         []any                             // Object data to be inserted in the graph
	mu              sync.Mutex                        // Mutex protecting access to the inserts array
	consumerChan    chan batchItem                    // Channel consuming inserts for async writing
	writingInFlight *sync.WaitGroup                   // Wait group tracking current unfinished writes
	batchSize       int                               // Batchsize of graph DB inserts
	qcounter        int32                             // Track items queued
	wcounter        int32                             // Track items writtn
	tags            []string                          // Telemetry tags
	writerTimeout   time.Duration                     // Timeout for the writer
	maxRetry        int                               // Maximum number of retries for failed writes
}

// NewJanusGraphAsyncEdgeWriter creates a new bulk edge writer instance.
func NewJanusGraphAsyncEdgeWriter(ctx context.Context, drc *gremlingo.DriverRemoteConnection,
	e edge.Builder, opts ...WriterOption,
) (*JanusGraphEdgeWriter, error) {
	options := &writerOptions{
		WriterTimeout: defaultWriterTimeout,
		MaxRetry:      defaultMaxRetry,
	}
	for _, opt := range opts {
		opt(options)
	}

	builder := fmt.Sprintf("%s::%s", e.Name(), e.Label())
	jw := JanusGraphEdgeWriter{
		builder:         builder,
		gremlin:         e.Traversal(),
		drc:             drc,
		inserts:         make([]any, 0, e.BatchSize()),
		traversalSource: gremlingo.Traversal_().WithRemote(drc),
		batchSize:       e.BatchSize(),
		writingInFlight: &sync.WaitGroup{},
		consumerChan:    make(chan batchItem, e.BatchSize()*channelSizeBatchFactor),
		tags:            append(options.Tags, tag.Label(e.Label()), tag.Builder(builder)),
		writerTimeout:   options.WriterTimeout,
		maxRetry:        options.MaxRetry,
	}

	jw.startBackgroundWriter(ctx)

	return &jw, nil
}

// startBackgroundWriter starts a background go routine
func (jgv *JanusGraphEdgeWriter) startBackgroundWriter(ctx context.Context) {
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
					log.Trace(ctx).Warn("Empty edge batch received in background janusgraph worker, skipping")
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

					log.Trace(ctx).Errorf("write data in background batch writer: %v", err)
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
func (jgv *JanusGraphEdgeWriter) retrySplitAndRequeue(ctx context.Context, batch *batchItem, e *errBatchWriter) {
	_ = statsd.Count(ctx, metric.RetryWriterCall, 1, jgv.tags, 1)

	// Compute the new batch size.
	newBatchSize := len(batch.data) / 2
	batch.retryCount++

	log.Trace(ctx).Warnf("Retrying write operation with smaller edge batch (n:%d -> %d, r:%d): %v", len(batch.data), newBatchSize, batch.retryCount, e.Unwrap())

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

// batchWrite will write a batch of entries into the graph DB and block until the write completes.
// Callers are responsible for doing an Add(1) to the writingInFlight wait group to ensure proper synchronization.
func (jgv *JanusGraphEdgeWriter) batchWrite(ctx context.Context, data []any) error {
	span, ctx := span.SpanRunFromContext(ctx, span.JanusGraphBatchWrite)
	span.SetTag(tag.LabelTag, jgv.builder)
	var err error
	defer func() { span.Finish(tracer.WithError(err)) }()
	defer jgv.writingInFlight.Done()

	datalen := len(data)
	_ = statsd.Count(ctx, metric.EdgeWrite, int64(datalen), jgv.tags, 1)
	log.Trace(ctx).Debugf("Batch write JanusGraphEdgeWriter with %d elements", datalen)
	atomic.AddInt32(&jgv.wcounter, int32(datalen)) //nolint:gosec // disable G115

	op := jgv.gremlin(jgv.traversalSource, data)
	promise := op.Iterate()

	// Wait for the write operation to complete or timeout.
	select {
	case <-ctx.Done():
		// If the context is cancelled, return the error.
		return ctx.Err()
	case <-time.After(jgv.writerTimeout):
		// If the write operation takes too long, return an error.
		return &errBatchWriter{
			err:       errors.New("edge write operation timed out"),
			retryable: true,
		}
	case err := <-promise:
		if err != nil {
			return fmt.Errorf("%s edge insert: %w", jgv.builder, err)
		}
	}

	return nil
}

func (jgv *JanusGraphEdgeWriter) Close(ctx context.Context) error {
	close(jgv.consumerChan)

	return nil
}

// Flush triggers writes of any remaining items in the queue.
// This is blocking
func (jgv *JanusGraphEdgeWriter) Flush(ctx context.Context) error {
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

	log.Trace(ctx).Debugf("Edge writer %d %s queued", jgv.qcounter, jgv.builder)
	log.Trace(ctx).Infof("Edge writer %d %s written", jgv.wcounter, jgv.builder)

	return nil
}

func (jgv *JanusGraphEdgeWriter) Queue(ctx context.Context, v any) error {
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
