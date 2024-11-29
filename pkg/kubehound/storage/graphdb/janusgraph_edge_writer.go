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
	writingInFlight *sync.WaitGroup                   // Wait group tracking current unfinished writes
	qcounter        int32                             // Track items queued
	wcounter        int32                             // Track items writtn
	tags            []string                          // Telemetry tags
	writerTimeout   time.Duration                     // Timeout for the writer
	maxRetry        int                               // Maximum number of retries for failed writes
	mb              *microBatcher                     // Micro batcher to batch writes
}

// NewJanusGraphAsyncEdgeWriter creates a new bulk edge writer instance.
func NewJanusGraphAsyncEdgeWriter(ctx context.Context, drc *gremlingo.DriverRemoteConnection,
	e edge.Builder, opts ...WriterOption,
) (*JanusGraphEdgeWriter, error) {
	options := &writerOptions{
		WriterTimeout:     defaultWriterTimeout,
		MaxRetry:          defaultMaxRetry,
		WriterWorkerCount: defaultWriterWorkerCount,
	}
	for _, opt := range opts {
		opt(options)
	}

	builder := fmt.Sprintf("%s::%s", e.Name(), e.Label())
	jw := JanusGraphEdgeWriter{
		builder:         builder,
		gremlin:         e.Traversal(),
		drc:             drc,
		traversalSource: gremlingo.Traversal_().WithRemote(drc),
		writingInFlight: &sync.WaitGroup{},
		tags:            append(options.Tags, tag.Label(e.Label()), tag.Builder(builder)),
		writerTimeout:   options.WriterTimeout,
		maxRetry:        options.MaxRetry,
	}

	// Create a new micro batcher to batch the inserts with split and retry logic.
	jw.mb = newMicroBatcher(log.Trace(ctx), e.BatchSize(), options.WriterWorkerCount, func(ctx context.Context, a []any) error {
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

// retrySplitAndRequeue will split the batch into smaller chunks and requeue them for writing.
func (jgv *JanusGraphEdgeWriter) splitAndRetry(ctx context.Context, retryCount int, payload []any) error {
	_ = statsd.Count(ctx, metric.RetryWriterCall, 1, jgv.tags, 1)

	// If we have reached the maximum number of retries, return an error.
	if retryCount >= jgv.maxRetry {
		return fmt.Errorf("max retry count reached: %d", retryCount)
	}

	// Compute the new batch size.
	newBatchSize := len(payload) / 2

	log.Trace(ctx).Warnf("Retrying write operation with smaller edge batch (n:%d -> %d, r:%d)", len(payload), newBatchSize, retryCount)

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

// batchWrite will write a batch of entries into the graph DB and block until the write completes.
func (jgv *JanusGraphEdgeWriter) batchWrite(ctx context.Context, data []any) error {
	span, ctx := span.SpanRunFromContext(ctx, span.JanusGraphBatchWrite)
	span.SetTag(tag.LabelTag, jgv.builder)
	var err error
	defer func() { span.Finish(tracer.WithError(err)) }()

	// Increment the writingInFlight wait group to track the number of writes in progress.
	jgv.writingInFlight.Add(1)
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
		return &batchWriterError{
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
	return nil
}

// Flush triggers writes of any remaining items in the queue.
// This is blocking
func (jgv *JanusGraphEdgeWriter) Flush(ctx context.Context) error {
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

	log.Trace(ctx).Debugf("Edge writer %d %s queued", jgv.qcounter, jgv.builder)
	log.Trace(ctx).Infof("Edge writer %d %s written", jgv.wcounter, jgv.builder)

	return nil
}

func (jgv *JanusGraphEdgeWriter) Queue(ctx context.Context, v any) error {
	atomic.AddInt32(&jgv.qcounter, 1)
	return jgv.mb.Enqueue(ctx, v)
}
