package graphdb

import (
	"context"

	"github.com/DataDog/KubeHound/pkg/globals"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/edge"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/vertex"
	gremlingo "github.com/apache/tinkerpop/gremlin-go/driver"
)

var _ Provider = (*JanusGraphProvider)(nil)

type JanusGraphProvider struct {
	client *gremlingo.DriverRemoteConnection
}

func NewGraphDriver(ctx context.Context, dbHost string) (*JanusGraphProvider, error) {
	driver, err := gremlingo.NewDriverRemoteConnection(dbHost)
	if err != nil {
		return nil, err
	}

	g := &JanusGraphProvider{
		client: driver,
	}

	return g, nil
}

func (jgp *JanusGraphProvider) Name() string {
	return "JanusGraphProvider"
}

func (jgp *JanusGraphProvider) HealthCheck(ctx context.Context) (bool, error) {
	return true, globals.ErrNotImplemented
}

// Raw returns a handle to the underlying provider to allow implementation specific operations e.g graph queries.
func (jgp *JanusGraphProvider) Raw() any {
	return jgp.client
}

// VertexWriter creates a new AsyncVertexWriter instance to enable asynchronous bulk inserts of vertices.
func (jgp *JanusGraphProvider) VertexWriter(ctx context.Context, v vertex.Vertex, opts ...WriterOption) (AsyncVertexWriter, error) {
	writer, err := NewJanusGraphAsyncVertexWriter(jgp.client, opts...)
	if err != nil {
		return nil, err
	}
	return writer, nil
}

// EdgeWriter creates a new AsyncEdgeWriter instance to enable asynchronous bulk inserts of edges.
func (jgp *JanusGraphProvider) EdgeWriter(ctx context.Context, e edge.Edge, opts ...WriterOption) (AsyncEdgeWriter, error) {
	writer, err := NewJanusGraphAsyncEdgeWriter(jgp.client)
	if err != nil {
		return nil, err
	}
	return writer, nil
}

// Close cleans up any resources used by the Provider implementation. Provider cannot be reused after this call.
func (jgp *JanusGraphProvider) Close(ctx context.Context) error {
	// This only logs errors
	jgp.client.Close()
	return nil
}