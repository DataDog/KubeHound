package graphdb

import (
	"context"
	"errors"
	"sync"

	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	gremlingo "github.com/apache/tinkerpop/gremlin-go/driver"
)

var _ AsyncVertexWriter = (*JanusGraphAsyncVertexWriter)(nil)

type GremlinTraversalVertex func(*gremlingo.GraphTraversalSource, []any) *gremlingo.GraphTraversal
type GremlinTraversalEdge func(*gremlingo.GraphTraversalSource, []any) *gremlingo.GraphTraversal

type JanusGraphAsyncVertexWriter struct {
	gremlin         GremlinTraversalVertex
	transaction     *gremlingo.Transaction
	traversalSource *gremlingo.GraphTraversalSource
	inserts         []any
	consumerChan    chan []any
	writingInFligth sync.WaitGroup
	batchSize       int // Shouldn't this be "per vertex types" ?
}

var _ AsyncEdgeWriter = (*JanusGraphAsyncEdgeWriter)(nil)

type JanusGraphAsyncEdgeWriter struct {
	gremlin         GremlinTraversalEdge
	transaction     *gremlingo.Transaction
	traversalSource *gremlingo.GraphTraversalSource
	inserts         []any
	consumerChan    chan []any
	writingInFligth sync.WaitGroup
	batchSize       int // Shouldn't this be "per edge types" ?
}

func NewJanusGraphAsyncEdgeWriter(drc *gremlingo.DriverRemoteConnection, opts ...WriterOption) (*JanusGraphAsyncEdgeWriter, error) {
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
		inserts:         make([]any, 0),
		transaction:     tx,
		traversalSource: gtx,
		batchSize:       1,
		consumerChan:    make(chan []any, 10),
	}

	return &jw, nil
}

func NewJanusGraphAsyncVertexWriter(drc *gremlingo.DriverRemoteConnection, opts ...WriterOption) (*JanusGraphAsyncVertexWriter, error) {
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
	jw := JanusGraphAsyncVertexWriter{
		inserts:         make([]interface{}, 0),
		transaction:     tx,
		traversalSource: gtx,
		batchSize:       1,
		consumerChan:    make(chan []any, 10),
	}

	return &jw, nil
}

func (jgv *JanusGraphAsyncVertexWriter) batchWrite(ctx context.Context, data []any) error {
	jgv.writingInFligth.Add(1)
	defer jgv.writingInFligth.Done()

	op := jgv.gremlin(jgv.traversalSource, data)
	promise := op.Iterate()
	err := <-promise
	if err != nil {
		jgv.transaction.Rollback()
		return err
	}

	return nil
}

func (jge *JanusGraphAsyncEdgeWriter) batchWrite(ctx context.Context, data []any) error {
	jge.writingInFligth.Add(1)
	defer jge.writingInFligth.Done()

	op := jge.gremlin(jge.traversalSource, data)
	promise := op.Iterate()
	err := <-promise
	if err != nil {
		jge.transaction.Rollback()
		return err
	}

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
