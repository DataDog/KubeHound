package graphdb

import (
	"context"
	"time"

	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/edge"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/vertex"
	"github.com/DataDog/KubeHound/pkg/kubehound/services"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/cache"
)

const (
	defaultWriterTimeout     = 60 * time.Second
	defaultMaxRetry          = 3
	defaultWriterWorkerCount = 10
)

type writerOptions struct {
	Tags              []string
	WriterWorkerCount int
	WriterTimeout     time.Duration
	MaxRetry          int
}

type WriterOption func(*writerOptions)

func WithTags(tags []string) WriterOption {
	return func(wo *writerOptions) {
		wo.Tags = append(wo.Tags, tags...)
	}
}

// WithWriterTimeout sets the timeout for the writer to complete the write operation.
func WithWriterTimeout(timeout time.Duration) WriterOption {
	return func(wo *writerOptions) {
		wo.WriterTimeout = timeout
	}
}

// WithWriterMaxRetry sets the maximum number of retries for failed writes.
func WithWriterMaxRetry(maxRetry int) WriterOption {
	return func(wo *writerOptions) {
		wo.MaxRetry = maxRetry
	}
}

// WithWriterWorkerCount sets the number of workers to process the batch.
func WithWriterWorkerCount(workerCount int) WriterOption {
	return func(wo *writerOptions) {
		wo.WriterWorkerCount = workerCount
	}
}

// Provider defines the interface for implementations of the graphdb provider for storage of the calculated K8s attack graph.
//
//go:generate mockery --name Provider --output mocks --case underscore --filename graph_provider.go --with-expecter
type Provider interface {
	services.Dependency

	// Prepare wipes all data from the graph (usually to ensure a clean start)
	Prepare(ctx context.Context) error

	// Raw returns a handle to the underlying provider to allow implementation specific operations e.g graph queries.
	Raw() any

	// Droping all assets from the graph database from a cluster name
	Clean(ctx context.Context, cluster string) error

	// VertexWriter creates a new AsyncVertexWriter instance to enable asynchronous bulk inserts of vertices.
	VertexWriter(ctx context.Context, v vertex.Builder, c cache.CacheProvider, opts ...WriterOption) (AsyncVertexWriter, error)

	// EdgeWriter creates a new AsyncEdgeWriter instance to enable asynchronous bulk inserts of edges.
	EdgeWriter(ctx context.Context, e edge.Builder, opts ...WriterOption) (AsyncEdgeWriter, error)

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

// Factory returns an initialized instance of a graphdb provider from the provided application config.
func Factory(ctx context.Context, cfg *config.KubehoundConfig) (Provider, error) {
	r := storage.Retrier(NewGraphDriver, cfg.Storage.Retry, cfg.Storage.RetryDelay)

	return r(ctx, cfg)
}
