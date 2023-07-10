package types

import (
	gremlin "github.com/apache/tinkerpop/gremlin-go/v3/driver"
)

// EdgeTraversal returns the function to create a graph database edge insert from an array of input objects.
type EdgeTraversal func(source *gremlin.GraphTraversalSource, inserts []TraversalInput) *gremlin.GraphTraversal

// VertexTraversal returns the function to create a graph database vertex insert from an array of input objects.
type VertexTraversal func(source *gremlin.GraphTraversalSource, inserts []TraversalInput) *gremlin.GraphTraversal
