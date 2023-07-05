package graphdb

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/DataDog/KubeHound/pkg/kubehound/graph/edge"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/types"
	"github.com/DataDog/KubeHound/pkg/telemetry"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	"github.com/DataDog/KubeHound/pkg/telemetry/statsd"
	gremlingo "github.com/apache/tinkerpop/gremlin-go/v3/driver"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

var _ AsyncEdgeWriter = (*JanusGraphEdgeWriter)(nil)

type JanusGraphEdgeWriter struct {
	builder         string                            // Qualified name of the edge being written
	gremlin         types.EdgeTraversal               // Gremlin traversal generator function
	drc             *gremlingo.DriverRemoteConnection // Gremlin driver remote connection
	traversalSource *gremlingo.GraphTraversalSource   // Transacted graph traversal source
	inserts         []types.TraversalInput            // Object data to be inserted in the graph
	mu              sync.Mutex                        // Mutex protecting access to the inserts array
	consumerChan    chan []types.TraversalInput       // Channel consuming inserts for async writing
	writingInFlight *sync.WaitGroup                   // Wait group tracking current unfinished writes
	batchSize       int                               // Batchsize of graph DB inserts
	qcounter        int32                             // Track items queued
	wcounter        int32                             // Track items writtn
	tags            []string                          // Telemetry tags
}

// NewJanusGraphAsyncEdgeWriter creates a new bulk edge writer instance.
func NewJanusGraphAsyncEdgeWriter(ctx context.Context, drc *gremlingo.DriverRemoteConnection,
	e edge.Builder, opts ...WriterOption) (*JanusGraphEdgeWriter, error) {

	options := &writerOptions{}
	for _, opt := range opts {
		opt(options)
	}

	jw := JanusGraphEdgeWriter{
		builder:         fmt.Sprintf("%s::%s", e.Name(), e.Label()),
		gremlin:         e.Traversal(),
		drc:             drc,
		inserts:         make([]types.TraversalInput, 0, e.BatchSize()),
		traversalSource: gremlingo.Traversal_().WithRemote(drc),
		batchSize:       e.BatchSize(),
		writingInFlight: &sync.WaitGroup{},
		consumerChan:    make(chan []types.TraversalInput, e.BatchSize()*channelSizeBatchFactor),
		tags:            options.Tags,
	}

	jw.startBackgroundWriter(ctx)

	return &jw, nil
}

// startBackgroundWriter starts a background go routine
func (jgv *JanusGraphEdgeWriter) startBackgroundWriter(ctx context.Context) {
	go func() {
		for {
			select {
			case data := <-jgv.consumerChan:
				// closing the channel shoud stop the go routine
				if data == nil {
					return
				}
				_ = statsd.Count(telemetry.MetricGraphdbBackgroundWriterCall, 1, jgv.tags, 1)
				err := jgv.batchWrite(ctx, data)
				if err != nil {
					log.Trace(ctx).Errorf("write data in background batch writer: %v", err)
				}
			case <-ctx.Done():
				log.Trace(ctx).Info("Closed background janusgraph worker on context cancel")
				return
			}
		}
	}()
}

// batchWrite will write a batch of entries into the graph DB and block until the write completes.
// Callers are responsible for doing an Add(1) to the writingInFlight wait group to ensure proper synchronization.
func (jgv *JanusGraphEdgeWriter) batchWrite(ctx context.Context, data []types.TraversalInput) error {
	span, ctx := tracer.StartSpanFromContext(ctx, telemetry.SpanJanusGraphOperationBatchWrite, tracer.Measured())
	span.SetTag(telemetry.TagKeyLabel, jgv.builder)
	defer span.Finish()
	defer jgv.writingInFlight.Done()

	tx := jgv.traversalSource.Tx()
	gtx, err := tx.Begin()
	if err != nil {
		return fmt.Errorf("%s edge insert transaction create: %w", jgv.builder, err)
	}
	defer tx.Close()

	datalen := len(data)
	_ = statsd.Gauge(telemetry.MetricGraphdbBatchWrite, float64(datalen), jgv.tags, 1)
	log.Trace(ctx).Debugf("Batch write JanusGraphEdgeWriter with %d elements", datalen)
	atomic.AddInt32(&jgv.wcounter, int32(datalen))

	op := jgv.gremlin(gtx, data)
	promise := op.Iterate()
	err = <-promise
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("%s vertex insert: %w", jgv.builder, err)
	}

	return tx.Commit()
}

func (jgv *JanusGraphEdgeWriter) Close(ctx context.Context) error {
	close(jgv.consumerChan)
	return nil
}

// Flush triggers writes of any remaining items in the queue.
// This is blocking
func (jgv *JanusGraphEdgeWriter) Flush(ctx context.Context) error {
	span, ctx := tracer.StartSpanFromContext(ctx, telemetry.SpanJanusGraphOperationFlush, tracer.Measured())
	span.SetTag(telemetry.TagKeyLabel, jgv.builder)
	defer span.Finish()

	jgv.mu.Lock()
	defer jgv.mu.Unlock()

	if jgv.traversalSource == nil {
		return errors.New("janusGraph traversalSource is not initialized")
	}

	if len(jgv.inserts) != 0 {
		jgv.writingInFlight.Add(1)
		err := jgv.batchWrite(ctx, jgv.inserts)
		if err != nil {
			log.Trace(ctx).Errorf("batch write %s: %+v", jgv.builder, err)
			jgv.writingInFlight.Wait()
			return err
		}

		log.Trace(ctx).Infof("Done flushing %s writes. clearing the queue", jgv.builder)
		jgv.inserts = nil
	}

	jgv.writingInFlight.Wait()

	log.Trace(ctx).Infof("Edge writer %d %s queued", jgv.qcounter, jgv.builder)
	log.Trace(ctx).Infof("Edge writer %d %s written", jgv.wcounter, jgv.builder)
	return nil
}

func (jgv *JanusGraphEdgeWriter) Queue(ctx context.Context, v any) error {
	jgv.mu.Lock()
	defer jgv.mu.Unlock()

	atomic.AddInt32(&jgv.qcounter, 1)
	jgv.inserts = append(jgv.inserts, v)

	_ = statsd.Gauge(telemetry.MetricGraphdbQueueSize, float64(len(jgv.inserts)), jgv.tags, 1)

	if len(jgv.inserts) > jgv.batchSize {
		copied := make([]types.TraversalInput, len(jgv.inserts))
		copy(copied, jgv.inserts)

		jgv.writingInFlight.Add(1)
		jgv.consumerChan <- copied

		// cleanup the ops array after we have copied it to the channel
		jgv.inserts = nil
	}

	return nil
}
