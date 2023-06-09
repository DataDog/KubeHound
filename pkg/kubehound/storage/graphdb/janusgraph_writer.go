package graphdb

import (
	"context"
	"errors"
	"sync"

	"github.com/DataDog/KubeHound/pkg/kubehound/graph/edge"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/vertex"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	gremlingo "github.com/apache/tinkerpop/gremlin-go/driver"
)

var _ AsyncVertexWriter = (*JanusGraphAsyncVertexWriter)(nil)

// type GremlinTraversalVertex func(*gremlingo.GraphTraversalSource, []any) *gremlingo.GraphTraversal
// type GremlinTraversalEdge func(*gremlingo.GraphTraversalSource, []any) *gremlingo.GraphTraversal

type JanusGraphAsyncVertexWriter struct {
	gremlin         vertex.VertexTraversal
	transaction     *gremlingo.Transaction
	traversalSource *gremlingo.GraphTraversalSource
	inserts         []any
	consumerChan    chan []any
	writingInFligth sync.WaitGroup
	batchSize       int // Shouldn't this be "per vertex types" ?
}

var _ AsyncEdgeWriter = (*JanusGraphAsyncEdgeWriter)(nil)

type JanusGraphAsyncEdgeWriter struct {
	gremlin         edge.EdgeTraversal
	transaction     *gremlingo.Transaction
	traversalSource *gremlingo.GraphTraversalSource
	inserts         []any
	consumerChan    chan []any
	writingInFligth sync.WaitGroup
	batchSize       int // Shouldn't this be "per edge types" ?
}

func NewJanusGraphAsyncEdgeWriter(drc *gremlingo.DriverRemoteConnection, e edge.Builder, opts ...WriterOption) (*JanusGraphAsyncEdgeWriter, error) {
	log.I.Infof("Created new JanusGraphAsyncEdgeWriter")
	options := &writerOptions{}
	for _, opt := range opts {
		opt(options)
	}

	traversal := gremlingo.Traversal_().WithRemote(drc)
	tx := traversal.Tx()
	gtx, err := tx.Begin()
	if err != nil {
		return nil, err
	}
	jw := JanusGraphAsyncEdgeWriter{
		gremlin:         e.Traversal(),
		inserts:         make([]any, 0),
		transaction:     tx,
		traversalSource: gtx,
		batchSize:       1,
		consumerChan:    make(chan []any, 10),
	}

	return &jw, nil
}

func NewJanusGraphAsyncVertexWriter(drc *gremlingo.DriverRemoteConnection, v vertex.Builder, opts ...WriterOption) (*JanusGraphAsyncVertexWriter, error) {
	log.I.Infof("Created new JanusGraphAsyncVertexWriter")
	options := &writerOptions{}
	for _, opt := range opts {
		opt(options)
	}

	source := gremlingo.Traversal_().WithRemote(drc)
	tx := source.Tx()
	gtx, err := tx.Begin()
	if err != nil {
		return nil, err
	}
	jw := JanusGraphAsyncVertexWriter{
		gremlin:         v.Traversal(),
		inserts:         make([]interface{}, 0),
		transaction:     tx,
		traversalSource: gtx,
		batchSize:       1,
		consumerChan:    make(chan []any, 10),
	}

	return &jw, nil
}

func (jgv *JanusGraphAsyncVertexWriter) batchWrite(ctx context.Context, data []any) error {
	log.I.Infof("batch write JanusGraphAsyncVertexWriter")
	jgv.writingInFligth.Add(1)
	defer jgv.writingInFligth.Done()

	convertedToTraversalInput := make([]vertex.TraversalInput, 0)
	for _, d := range data {
		convertedToTraversalInput = append(convertedToTraversalInput, d.(vertex.TraversalInput))
	}

	log.I.Infof("BEFORE gremlin()")
	op := jgv.gremlin(jgv.traversalSource, convertedToTraversalInput)
	log.I.Infof("BEFORE ITERATE")
	promise := op.Iterate()
	log.I.Infof("BEFORE AFTER ITERATE")
	log.I.Infof("BEFORE PROMISE")
	err := <-promise
	log.I.Infof("AFTER PROMISE: %v, convertedToTraversalInput: %+v", err, convertedToTraversalInput)
	if err != nil {
		jgv.transaction.Rollback()
		return err
	}
	jgv.transaction.Commit()
	log.I.Infof("=== DONE == batch write JanusGraphAsyncVertexWriter")
	return nil
}

func (jge *JanusGraphAsyncEdgeWriter) batchWrite(ctx context.Context, data []any) error {
	log.I.Infof("batch write JanusGraphAsyncEdgeWriter")
	jge.writingInFligth.Add(1)
	defer jge.writingInFligth.Done()

	if jge.gremlin == nil {
		panic("lol")
	}

	// This seems ~pointless BUT is required to have the ability to use edge.TraversalInput/vertex.TraversalInput
	// as the type
	// Even tho it's an alias to any, since we use it in an array, we cannot simply .([]any) or vice versa because of the underlying memory layout.
	convertedToTraversalInput := make([]edge.TraversalInput, 0)
	for _, d := range data {
		convertedToTraversalInput = append(convertedToTraversalInput, d.(edge.TraversalInput))
	}

	op := jge.gremlin(jge.traversalSource, convertedToTraversalInput)
	promise := op.Iterate()
	err := <-promise
	if err != nil {
		jge.transaction.Rollback()
		return err
	}
	jge.transaction.Commit()
	log.I.Infof("=== DONE == batch write JanusGraphAsyncEdgeWriter")
	return nil
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
				log.I.Info("Closed background Janus Graph worker (vertex)")
				return
			}
		}
	}()
}

func (v *JanusGraphAsyncVertexWriter) Close(ctx context.Context) error {
	return v.transaction.Close()
}

func (e *JanusGraphAsyncEdgeWriter) Close(ctx context.Context) error {
	return e.transaction.Close()
}

func (v *JanusGraphAsyncVertexWriter) Flush(ctx context.Context) error {
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
	log.I.Infof("Flushing remaining of queue for vertices: %+v", v.inserts)
	err := v.batchWrite(ctx, v.inserts)
	if err != nil {
		v.writingInFligth.Wait()
		return err
	}
	v.inserts = nil

	return nil
}

func (e *JanusGraphAsyncEdgeWriter) Flush(ctx context.Context) error {
	if e.traversalSource == nil {
		return errors.New("JanusGraph traversalSource is not initialized")
	}

	if len(e.inserts) == 0 {
		log.I.Debugf("Skipping flush on edges writer as no write operations left")
		// we need to send something to the channel from this function whenever we don't return an error
		// we cannot defer it because the go routine may last longer than the current function
		// the defer is going to be executed at the return time, whetever or not the inner go routine is processing data
		e.writingInFligth.Wait()
		return nil
	}

	err := e.batchWrite(ctx, e.inserts)
	if err != nil {
		e.writingInFligth.Wait()
		return err
	}
	e.inserts = nil

	return nil
}

func (vw *JanusGraphAsyncVertexWriter) Queue(ctx context.Context, vertex any) error {
	if len(vw.inserts) > vw.batchSize {
		vw.consumerChan <- vw.inserts
		// cleanup the ops array after we have copied it to the channel
		vw.inserts = nil
	}
	vw.inserts = append(vw.inserts, vertex)
	return nil
}

func (e *JanusGraphAsyncEdgeWriter) Queue(ctx context.Context, edge any) error {
	if len(e.inserts) > e.batchSize {
		e.consumerChan <- e.inserts
		// cleanup the ops array after we have copied it to the channel
		e.inserts = nil
	}
	e.inserts = append(e.inserts, edge)
	return nil
}
