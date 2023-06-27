package graphdb

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"

	"github.com/DataDog/KubeHound/pkg/kubehound/graph/edge"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/path"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/vertex"
	"github.com/DataDog/KubeHound/pkg/telemetry"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	"github.com/DataDog/KubeHound/pkg/telemetry/statsd"
	gremlingo "github.com/apache/tinkerpop/gremlin-go/v3/driver"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

type TWriterInput interface {
	vertex.Traversal | edge.Traversal | path.Traversal
}

type JanusGraphAsyncWriter[T TWriterInput] struct {
	label           string                            // Label of the graph entity being written
	gremlin         T                                 // Gremlin traversal generator function
	drc             *gremlingo.DriverRemoteConnection // Gremlin driver remote connection
	traversalSource *gremlingo.GraphTraversalSource   // Transacted graph traversal source
	inserts         []types.TraversalInput            // Object data to be inserted in the graph
	mu              sync.Mutex                        // Mutex protecting access to the inserts array
	consumerChan    chan []types.TraversalInput       // Channel consuming inserts for async writing
	writingInFlight *sync.WaitGroup                   // Wait group tracking current unfinished writes
	batchSize       int                               // Batchsize of graph DB inserts
	qcounter        int32                             // Track items queued
	wcounter        int32                             // Track items writtn
	tags            []string
}

// startBackgroundWriter starts a background go routine
func (jgv *JanusGraphAsyncWriter[T]) startBackgroundWriter(ctx context.Context) {
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
					log.I.Errorf("write data in background batch writer: %v", err)
				}
			case <-ctx.Done():
				log.I.Info("Closed background janusgraph worker on context cancel")
				return
			}
		}
	}()
}

// batchWrite will write a batch of entries into the graph DB and block until the write completes.
// Callers are responsible for doing an Add(1) to the writingInFlight wait group to ensure proper synchronization.
func (jgv *JanusGraphAsyncWriter[T]) batchWrite(ctx context.Context, data []types.TraversalInput) error {
	span, _ := tracer.StartSpanFromContext(ctx, telemetry.SpanJanusGraphOperationBatchWrite, tracer.Measured())
	span.SetTag(telemetry.TagKeyLabel, jgv.label)
	defer span.Finish()

	tx := jgv.traversalSource.Tx()
	gtx, err := tx.Begin()
	if err != nil {
		return err
	}
	defer tx.Close()

	datalen := len(data)
	_ = statsd.Gauge(telemetry.MetricGraphdbBatchWrite, float64(datalen), jgv.tags, 1)

	log.I.Debugf("batch write JanusGraphAsyncVertexWriter with %d elements", datalen)
	defer jgv.writingInFlight.Done()

	atomic.AddInt32(&jgv.wcounter, int32(datalen))
	op := jgv.gremlin(gtx, data)
	promise := op.Iterate()
	err = <-promise
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (jgv *JanusGraphAsyncWriter[T]) Close(ctx context.Context) error {
	close(jgv.consumerChan)
	return nil
}

// Flush triggers writes of any remaining items in the queue.
// This is blocking
func (jgv *JanusGraphAsyncWriter[T]) Flush(ctx context.Context) error {
	span, ctx := tracer.StartSpanFromContext(ctx, telemetry.SpanJanusGraphOperationFlush, tracer.Measured())
	span.SetTag(telemetry.TagKeyLabel, jgv.label)
	defer span.Finish()

	jgv.mu.Lock()
	defer jgv.mu.Unlock()

	if jgv.traversalSource == nil {
		return errors.New("JanusGraph traversalSource is not initialized")
	}

	if len(jgv.inserts) != 0 {
		jgv.writingInFlight.Add(1)
		err := jgv.batchWrite(ctx, jgv.inserts)
		if err != nil {
			log.I.Errorf("batch write %s: %+v", jgv.label, err)
			jgv.writingInFlight.Wait()
			return err
		}

		log.I.Infof("Done flushing %s writes. clearing the queue", jgv.label)
		jgv.inserts = nil
	}

	jgv.writingInFlight.Wait()

	// TODO replace with telemetry.metrics
	log.I.Infof("%d %s queued", jgv.qcounter, jgv.label)
	log.I.Infof("%d %s written", jgv.wcounter, jgv.label)
	return nil
}

func (jgv *JanusGraphAsyncWriter[T]) Queue(ctx context.Context, v any) error {
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
