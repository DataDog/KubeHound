package vertex

import (
	gremlin "github.com/apache/tinkerpop/gremlin-go/driver"
)

// An object to be consumed by a vertex traversal function to insert a vertex into the graph database.
type TraversalInput any

// VertexTraversal returns the function to create a graph database vertex insert from an array of input objects.
type VertexTraversal func(g *gremlin.GraphTraversal, insert TraversalInput) *gremlin.GraphTraversal

// Vertex interface defines objects used to construct vertices within our graph database through processing data from an ingestion pipeline.
type Vertex interface {
	// Label returns the label for the vertex (convention is all camelcase i.e VertexName)
	Label() string

	// Traversal returns a graph traversal function that enables creating vertices from an input array of TraversalInput objects.
	Traversal() VertexTraversal
}
