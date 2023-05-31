package graphdb

import (
	"context"

	"github.com/DataDog/KubeHound/pkg/globals"
	gremlingo "github.com/apache/tinkerpop/gremlin-go/driver"
)

var _ AsyncVertexWriter = (*JanusGraphAsyncVertexWriter)(nil)

type GremlinTraversal func(*gremlingo.GraphTraversalSource, []any) *gremlingo.GraphTraversal

type JanusGraphAsyncVertexWriter struct {
	gremlin     GremlinTraversal
	transaction *gremlingo.Transaction
	traversal   *gremlingo.GraphTraversalSource
	inserts     []any
}

func NewJanusGraphAsyncVertexWriter() *JanusGraphAsyncVertexWriter {
	jw := JanusGraphAsyncVertexWriter{}
	jw.inserts = make([]any, 0)
	return &jw
}

var _ AsyncEdgeWriter = (*JanusGraphAsyncEdgeWriter)(nil)

type JanusGraphAsyncEdgeWriter struct {
	gremlin     GremlinTraversal
	transaction *gremlingo.Transaction
	traversal   *gremlingo.GraphTraversalSource
	inserts     []any
}

func NewJanusGraphAsyncEdgeWriter() *JanusGraphAsyncEdgeWriter {
	jw := JanusGraphAsyncEdgeWriter{}
	jw.inserts = make([]any, 0)
	return &jw
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
	return globals.ErrNotImplemented
}

func (e *JanusGraphAsyncEdgeWriter) Queue(ctx context.Context, edge any) error {
	return globals.ErrNotImplemented
}
