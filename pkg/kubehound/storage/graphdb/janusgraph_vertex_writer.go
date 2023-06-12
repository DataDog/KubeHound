package graphdb

import (
	"context"
	"errors"
	"sync"

	"github.com/DataDog/KubeHound/pkg/kubehound/graph/vertex"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	gremlingo "github.com/apache/tinkerpop/gremlin-go/driver"
)

var _ AsyncVertexWriter = (*JanusGraphAsyncVertexWriter)(nil)

// type GremlinTraversalVertex func(*gremlingo.GraphTraversalSource, []any) *gremlingo.GraphTraversal
// type GremlinTraversalEdge func(*gremlingo.GraphTraversalSource, []any) *gremlingo.GraphTraversal

type JanusGraphAsyncVertexWriter struct {
	gremlin              vertex.VertexTraversal
	transaction          *gremlingo.Transaction
	traversalSource      *gremlingo.GraphTraversalSource
	inserts              []any
	consumerChan         chan []any
	writingInFligth      sync.WaitGroup
	batchSize            int
	mu                   sync.Mutex
	isTransactionEnabled bool
}

func NewJanusGraphAsyncVertexWriter(ctx context.Context, drc *gremlingo.DriverRemoteConnection, v vertex.Builder, opts ...WriterOption) (*JanusGraphAsyncVertexWriter, error) {
	log.I.Infof("Created new JanusGraphAsyncVertexWriter")
	options := &writerOptions{}
	for _, opt := range opts {
		opt(options)
	}

	source := gremlingo.Traversal_().WithRemote(drc)
	// quick switch to enable / disable transaction
	var tx *gremlingo.Transaction
	if options.isTransactionEnabled {
		log.I.Info("GraphDB transaction enabled!")
		tx = source.Tx()
		var err error
		source, err = tx.Begin()
		if err != nil {
			return nil, err
		}
	}

	jw := JanusGraphAsyncVertexWriter{
		gremlin:         v.Traversal(),
		inserts:         make([]interface{}, 0),
		transaction:     tx,
		traversalSource: source,
		batchSize:       v.BatchSize(),
		consumerChan:    make(chan []any, v.BatchSize()*channelSizeBatchFactor),
	}
	jw.backgroundWriter(ctx)
	return &jw, nil
}

// backgroundWriter starts a background go routine
func (jgv *JanusGraphAsyncVertexWriter) backgroundWriter(ctx context.Context) {
	go func() {
		for {
			select {
			case data := <-jgv.consumerChan:
				// closing the channel shoud stop the go routine
				if data == nil {
					return
				}
				err := jgv.batchWrite(ctx, data)
				if err != nil {
					log.I.Errorf("failed to write data in background batch writer: %v", err)
				}
			case <-ctx.Done():
				log.I.Info("Closed background janusgraph worker")
				return
			}
		}
	}()
}

func (jgv *JanusGraphAsyncVertexWriter) batchWrite(ctx context.Context, data []any) error {
	log.I.Debugf("batch write JanusGraphAsyncVertexWriter with %d elements", len(data))
	jgv.writingInFligth.Add(1)
	defer jgv.writingInFligth.Done()

	convertedToTraversalInput := make([]vertex.TraversalInput, 0)
	for _, d := range data {
		convertedToTraversalInput = append(convertedToTraversalInput, d.(vertex.TraversalInput))
	}
	op := jgv.gremlin(jgv.traversalSource, convertedToTraversalInput)
	promise := op.Iterate()
	err := <-promise
	if err != nil {
		if jgv.isTransactionEnabled {
			jgv.transaction.Rollback()
		}
		return err
	}
	if jgv.isTransactionEnabled {
		log.I.Infof("commiting work")
		err = jgv.transaction.Commit()
		if err != nil {
			log.I.Errorf("failed to commit: %+v", err)
			return err
		}
	}
	return nil
}

func (v *JanusGraphAsyncVertexWriter) Close(ctx context.Context) error {
	if v.isTransactionEnabled {
		return v.transaction.Close()
	}
	return nil
}

func (v *JanusGraphAsyncVertexWriter) Flush(ctx context.Context) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	if v.traversalSource == nil {
		return errors.New("JanusGraph traversalSource is not initialized")
	}

	if len(v.inserts) == 0 {
		log.I.Debugf("Skipping flush on vertex writer as no write operations left")
		// we need to send something to the channel from this function whenever we don't return an error
		// we cannot defer it because the go routine may last longer than the current function
		// the defer is going to be executed at the return time, whetever or not the inner go routine is processing data
		v.writingInFligth.Wait()
		return nil
	}

	err := v.batchWrite(ctx, v.inserts)
	if err != nil {
		log.I.Errorf("Failed to batch write vertex: %+v", err)
		v.writingInFligth.Wait()
		return err
	}
	log.I.Info("Done flushing vertices, clearing the queue")
	v.inserts = nil

	return nil
}

func (vw *JanusGraphAsyncVertexWriter) Queue(ctx context.Context, vertex any) error {
	vw.mu.Lock()
	defer vw.mu.Unlock()
	if len(vw.inserts) > vw.batchSize {
		copied := make([]any, len(vw.inserts))
		copy(copied, vw.inserts)
		vw.consumerChan <- copied
		// cleanup the ops array after we have copied it to the channel
		vw.inserts = nil
	}
	vw.inserts = append(vw.inserts, vertex)
	return nil
}
