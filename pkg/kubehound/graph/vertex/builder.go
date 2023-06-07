package vertex

import (
	"context"

	"github.com/DataDog/KubeHound/pkg/kubehound/graph/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/cache"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/storedb"
	gremlin "github.com/apache/tinkerpop/gremlin-go/driver"
)

const (
	DefaultBatchSize = 100
)

// VertexTraversal returns the function to create a graph database vertex insert from an array of input objects.
type VertexTraversal func(source *gremlin.GraphTraversalSource, inserts []types.TraversalInput) *gremlin.GraphTraversal

// PipelineBuilder interface defines objects used to construct vertices within our graph database through processing data from an ingestion pipeline.
type PipelineBuilder interface {
	// Label returns the label for the vertex (convention is all camelcase i.e VertexName)
	Label() string

	// BatchSize returns the batch size of bulk inserts (and threshold for triggering a flush).
	BatchSize() int

	// Traversal returns a graph traversal function that enables creating vertices from an input array of TraversalInput objects.
	Traversal() VertexTraversal
}

// PipelineBuilder interface defines objects used to construct vertices within our graph database AFTER the ingestion pipeline has completed.
type QueryBuilder interface {
	PipelineBuilder

	// Stream will query the store db for the data required to create an edge and stream to graph DB via callbacks.
	// Each query result is encapsulated within an DataContainer and transformed to a TraversalInput via a call to
	// the edge's Processor function. Invoking the complete callback signals the end of the stream.
	Stream(ctx context.Context, sdb storedb.Provider, cache cache.CacheReader,
		process types.ProcessEntryCallback, complete types.CompleteQueryCallback) error

	// Processor translates an DataContainer retrieved from the data store into a TraversalInput to pass to the traversal.
	Processor(ctx context.Context, model types.DataContainer) (types.TraversalInput, error)
}
