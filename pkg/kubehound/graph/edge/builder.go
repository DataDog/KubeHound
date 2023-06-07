package edge

import (
	"context"

	"github.com/DataDog/KubeHound/pkg/kubehound/graph/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/cache"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/storedb"
	gremlin "github.com/apache/tinkerpop/gremlin-go/driver"
)

const (
	DefaultBatchSize = 200
)

// EdgeTraversal returns the function to create a graph database edge insert from an array of input objects.
type EdgeTraversal func(source *gremlin.GraphTraversalSource, inserts []types.TraversalInput) *gremlin.GraphTraversal

// Edge interface defines objects used to construct edges within our graph database through processing data from the intermediate store.

//go:generate mockery --name Builder --output mocks --case underscore --filename edge.go --with-expecter
type Builder interface {
	// Label returns the label for the edge (convention is all uppercase i.e EDGE_NAME)
	Label() string

	// BatchSize returns the batch size of bulk inserts (and threshold for triggering a flush).
	BatchSize() int

	// Traversal returns a graph traversal function that enables creating edges from an input array of TraversalInput objects.
	Traversal() EdgeTraversal

	// Stream will query the store db for the data required to create an edge and stream to graph DB via callbacks.
	// Each query result is encapsulated within an DataContainer and transformed to a TraversalInput via a call to
	// the edge's Processor function. Invoking the complete callback signals the end of the stream.
	Stream(ctx context.Context, store storedb.Provider, cache cache.CacheReader,
		process types.ProcessEntryCallback, complete types.CompleteQueryCallback) error
}
