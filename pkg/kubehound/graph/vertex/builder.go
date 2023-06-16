package vertex

import (
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/types"
	gremlin "github.com/apache/tinkerpop/gremlin-go/v3/driver"
)

const (
	DefaultBatchSize = 50
)

// Traversal returns the function to create a graph database vertex insert from an array of input objects.
type Traversal func(source *gremlin.GraphTraversalSource, inserts []types.TraversalInput) *gremlin.GraphTraversal

// Builder interface defines objects used to construct vertices within our graph database through processing data from an ingestion pipeline.
type Builder interface {
	// Label returns the label for the vertex (convention is all camelcase i.e VertexName)
	Label() string

	// BatchSize returns the batch size of bulk inserts (and threshold for triggering a flush).
	BatchSize() int

	// Traversal returns a graph traversal function that enables creating vertices from an input array of TraversalInput objects.
	Traversal() Traversal
}
