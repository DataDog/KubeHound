package vertex

import (
	gremlin "github.com/apache/tinkerpop/gremlin-go/driver"
)

const (
	DefaultBatchSize = 50
)

// An object to be consumed by a vertex traversal function to insert a vertex into the graph database.
type TraversalInput any

// VertexTraversal returns the function to create a graph database vertex insert from an array of input objects.
type VertexTraversal func(source *gremlin.GraphTraversalSource, inserts []TraversalInput) *gremlin.GraphTraversal

// Builder interface defines objects used to construct vertices within our graph database through processing data from an ingestion pipeline.
type Builder interface {
	// Label returns the label for the vertex (convention is all camelcase i.e VertexName)
	Label() string

	// BatchSize returns the batch size of bulk inserts (and threshold for triggering a flush).
	BatchSize() int

	// Traversal returns a graph traversal function that enables creating vertices from an input array of TraversalInput objects.
	Traversal() VertexTraversal
}
