package graphdb

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/DataDog/KubeHound/pkg/kubehound/graph/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/vertex"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/cache"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/cache/cachekey"
	"github.com/DataDog/KubeHound/pkg/telemetry"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	"github.com/DataDog/KubeHound/pkg/telemetry/statsd"
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
	consumerChan    chan []any                      // Channel consuming inserts for async writing
	writingInFlight *sync.WaitGroup                 // Wait group tracking current unfinished writes
	batchSize       int                             // Batchsize of graph DB inserts
	qcounter        int32                           // Track items queued
	wcounter        int32                           // Track items writtn
	tags            []string                        // Telemetry tags
	cache           cache.AsyncWriter               // Cache writer to cache store id -> vertex id mappings
}

// NewJanusGraphAsyncVertexWriter creates a new bulk vertex writer instance.
func NewJanusGraphAsyncVertexWriter(ctx context.Context, drc *gremlin.DriverRemoteConnection,
	v vertex.Builder, c cache.CacheProvider, opts ...WriterOption) (*JanusGraphVertexWriter, error) {

	options := &writerOptions{}
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
		consumerChan:    make(chan []any, v.BatchSize()*channelSizeBatchFactor),
		tags:            append(options.Tags),
		cache:           cw,
	}

	jw.startBackgroundWriter(ctx)

	return &jw, nil
}

// startBackgroundWriter starts a background go routine
func (jgv *JanusGraphVertexWriter) startBackgroundWriter(ctx context.Context) {
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
					log.Trace(ctx).Errorf("Write data in background batch writer: %v", err)
				}
			case <-ctx.Done():
				log.Trace(ctx).Info("Closed background janusgraph worker on context cancel")

				return
			}
		}
	}()
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
	span, ctx := tracer.StartSpanFromContext(ctx, telemetry.SpanJanusGraphOperationBatchWrite,
		tracer.Measured(), tracer.ServiceName(TracerServicename))
	span.SetTag(telemetry.TagKeyLabel, jgv.builder)
	defer span.Finish()
	defer jgv.writingInFlight.Done()

	datalen := len(data)
	_ = statsd.Gauge(telemetry.MetricGraphdbBatchWrite, float64(datalen), jgv.tags, 1)
	log.Trace(ctx).Debugf("Batch write JanusGraphVertexWriter with %d elements", datalen)
	atomic.AddInt32(&jgv.wcounter, int32(datalen))

	op := jgv.gremlin(jgv.traversalSource, data)
	raw, err := op.Project("id", "storeID").
		By(gremlin.T.Id).
		By("storeID").
		ToList()
	if err != nil {
		return fmt.Errorf("%s vertex insert: %w", jgv.builder, err)
	}

	// Gremlin will return a list of maps containing and vertex id and store id values for each vertex inserted.
	// We need to parse each map entry and add to our cache.
	if err := jgv.cacheIds(ctx, raw); err != nil {
		return err
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
	span, ctx := tracer.StartSpanFromContext(ctx, telemetry.SpanJanusGraphOperationFlush,
		tracer.Measured(), tracer.ServiceName(TracerServicename))
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

		log.Trace(ctx).Debugf("Done flushing %s writes. clearing the queue", jgv.builder)
		jgv.inserts = nil
	}

	jgv.writingInFlight.Wait()

	err := jgv.cache.Flush(ctx)
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

	_ = statsd.Gauge(telemetry.MetricGraphdbQueueSize, float64(len(jgv.inserts)), jgv.tags, 1)

	if len(jgv.inserts) > jgv.batchSize {
		copied := make([]any, len(jgv.inserts))
		copy(copied, jgv.inserts)

		jgv.writingInFlight.Add(1)
		jgv.consumerChan <- copied

		// cleanup the ops array after we have copied it to the channel
		jgv.inserts = nil
	}

	return nil
}
