package path

import (
	"context"

	"github.com/DataDog/KubeHound/pkg/kubehound/graph/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/cache"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/storedb"
	gremlin "github.com/apache/tinkerpop/gremlin-go/v3/driver"
)

const (
	DefaultBatchSize = 25
)

// Optional syntactic sugar.
var __ = gremlin.T__

// Traversal returns the function to create a graph database path insert from an array of input objects.
// Paths are edges that result in one or more new vertex creation e.g TOKEN_BRUTEFORCE enables creation of a new Token vertex.
type Traversal func(source *gremlin.GraphTraversalSource, inserts []types.TraversalInput) *gremlin.GraphTraversal

// Builder interface defines objects used to construct paths within our graph database AFTER the ingestion pipeline has completed.
type Builder interface {
	// Label returns the label for the vertex (convention is all uppercase PATH_NAME)
	Label() string

	// BatchSize returns the batch size of bulk inserts (and threshold for triggering a flush).
	BatchSize() int

	// Traversal returns a graph traversal function that enables creating vertices from an input array of TraversalInput objects.
	Traversal() Traversal

	// Stream will query the store db for the data required to create an edge and stream to graph DB via callbacks.
	// Each query result is encapsulated within an DataContainer and transformed to a TraversalInput. Invoking the complete callback signals the end of the stream.
	Stream(ctx context.Context, sdb storedb.Provider, cache cache.CacheReader,
		process types.ProcessEntryCallback, complete types.CompleteQueryCallback) error
}
