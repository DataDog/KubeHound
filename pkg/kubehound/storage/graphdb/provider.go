package graphdb

import (
	"context"

	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/edge"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/path"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/vertex"
	"github.com/DataDog/KubeHound/pkg/kubehound/services"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage"
)

type writerOptions struct {
}

type WriterOption func(*writerOptions)

// Provider defines the interface for implementations of the graphdb provider for storage of the calculated K8s attack graph.
//
//go:generate mockery --name Provider --output mocks --case underscore --filename graph_provider.go --with-expecter
type Provider interface {
	services.Dependency

	// Raw returns a handle to the underlying provider to allow implementation specific operations e.g graph queries.
	Raw() any

	// VertexWriter creates a new AsyncVertexWriter instance to enable asynchronous bulk inserts of vertices.
	VertexWriter(ctx context.Context, v vertex.Builder, opts ...WriterOption) (AsyncVertexWriter, error)

	// EdgeWriter creates a new AsyncEdgeWriter instance to enable asynchronous bulk inserts of edges.
	EdgeWriter(ctx context.Context, e edge.Builder, opts ...WriterOption) (AsyncEdgeWriter, error)

	// PathWriter creates a new AsyncPathWriter instance to enable asynchronous bulk inserts of paths.
	PathWriter(ctx context.Context, p path.Builder, opts ...WriterOption) (AsyncPathWriter, error)

	// Close cleans up any resources used by the Provider implementation. Provider cannot be reused after this call.
	Close(ctx context.Context) error
}

type WriterBase interface {
	// Flush triggers writes of any remaining items in the queue. Blocks until operation completes.
	Flush(ctx context.Context) error

	// Close cleans up any resources used by the writer implementation. Writer cannot be reused after this call.
	Close(ctx context.Context) error
}

// AsyncVertexWriter defines the interface for writer clients to queue aysnchronous, batched writes  of vertices to the graphdb.
//
//go:generate mockery --name AsyncVertexWriter --output mocks --case underscore --filename vertex_writer.go --with-expecter
type AsyncVertexWriter interface {
	WriterBase

	// Queue add a vertex model to an asynchronous write queue. Non-blocking.
	Queue(ctx context.Context, v any) error
}

// AsyncEdgeWriter defines the interface for writer clients to queue aysnchronous, batched writes of edges to the graphdb.
//
//go:generate mockery --name AsyncEdgeWriter --output mocks --case underscore --filename edge_writer.go --with-expecter
type AsyncEdgeWriter interface {
	WriterBase

	// Queue add an edge model to an asynchronous write queue. Non-blocking.
	Queue(ctx context.Context, e any) error
}

// AsyncPathWriter defines the interface for writer clients to queue aysnchronous, batched writes of paths to the graphdb.
type AsyncPathWriter interface {
	WriterBase

	// Queue add a path model to an asynchronous write queue. Non-blocking.
	Queue(ctx context.Context, e any) error
}

// Factory returns an initialized instance of a graphdb provider from the provided application config.
func Factory(ctx context.Context, cfg *config.KubehoundConfig) (Provider, error) {
	r := storage.Retry(NewGraphDriver, cfg.Storage.Retry, cfg.Storage.RetryDelay)
	return r(ctx, cfg.JanusGraph.URL, cfg.JanusGraph.ConnectionTimeout)
}
