package vertex

import (
	gremlingo "github.com/apache/tinkerpop/gremlin-go/driver"
)

type VertexTraversal func(g *gremlingo.GraphTraversal, insert any) *gremlingo.GraphTraversal

type Vertex interface {
	Label() string
	Traversal() VertexTraversal
}
