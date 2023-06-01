package graphdb

import (
	"context"
	"sync"

	"github.com/DataDog/KubeHound/pkg/globals"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	gremlingo "github.com/apache/tinkerpop/gremlin-go/driver"
)

var _ AsyncVertexWriter = (*JanusGraphAsyncVertexWriter)(nil)

type GremlinTraversalVertex func(*gremlingo.GraphTraversalSource, []gremlingo.Vertex) *gremlingo.GraphTraversal
type GremlinTraversalEdge func(*gremlingo.GraphTraversalSource, []gremlingo.Edge) *gremlingo.GraphTraversal

type JanusGraphAsyncVertexWriter struct {
	gremlin         GremlinTraversalVertex
	transaction     *gremlingo.Transaction
	traversalSource *gremlingo.GraphTraversalSource
	inserts         []any
	consumerChan    chan []gremlingo.Vertex
	writingInFligth sync.WaitGroup
}

var _ AsyncEdgeWriter = (*JanusGraphAsyncEdgeWriter)(nil)

type JanusGraphAsyncEdgeWriter struct {
	gremlin         GremlinTraversalEdge
	transaction     *gremlingo.Transaction
	traversalSource *gremlingo.GraphTraversalSource
	inserts         []any
	writingInFligth sync.WaitGroup
}

func NewJanusGraphAsyncEdgeWriter() *JanusGraphAsyncEdgeWriter {
	jw := JanusGraphAsyncEdgeWriter{}
	jw.inserts = make([]any, 0)
	return &jw
}

func NewJanusGraphAsyncVertexWriter(drc *gremlingo.DriverRemoteConnection) (*JanusGraphAsyncVertexWriter, error) {
	traversal := gremlingo.Traversal_().WithRemote(drc)
	tx := traversal.Tx()
	gtx, err := tx.Begin()
	if err != nil {
		return nil, err
	}
	jw := JanusGraphAsyncVertexWriter{
		inserts:         make([]any, 0),
		transaction:     tx,
		traversalSource: gtx,
	}

	return &jw, nil
}

func (jgv *JanusGraphAsyncVertexWriter) batchWrite(ctx context.Context, data []gremlingo.Vertex) error {
	op := jgv.gremlin(jgv.traversalSource, data)
	promise := op.Iterate()
	err := <-promise
	if err != nil {
		jgv.transaction.Rollback()
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
					log.I.Errorf("failed to write data in background batch writer: %w", err)
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
	return globals.ErrNotImplemented
}

func (e *JanusGraphAsyncEdgeWriter) Flush(ctx context.Context) error {
	return globals.ErrNotImplemented
}

func (v *JanusGraphAsyncVertexWriter) Queue(ctx context.Context, vertex any) error {
	v.writingInFligth.Add(1)
	return globals.ErrNotImplemented
}

func (e *JanusGraphAsyncEdgeWriter) Queue(ctx context.Context, edge any) error {
	e.writingInFligth.Add(1)
	return globals.ErrNotImplemented
}
