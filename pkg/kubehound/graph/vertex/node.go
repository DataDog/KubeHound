package vertex

import (
	gremlin "github.com/apache/tinkerpop/gremlin-go/driver"
	gremlingo "github.com/apache/tinkerpop/gremlin-go/driver"
)

const (
	nodeLabel = "Node"
)

var _ Builder = (*Node)(nil)

type Node struct {
}

func (v Node) Label() string {
	return nodeLabel
}

func (v Node) BatchSize() int {
	return DefaultBatchSize
}

func (v Node) Traversal() VertexTraversal {
	return func(g *gremlin.GraphTraversalSource, inserts []TraversalInput) *gremlin.GraphTraversal {
		traversal := g.Inject(inserts).Unfold().As("c").
			AddV(v.Label()).
			Property("store_id", gremlingo.T__.Select("c").Select("store_id")).
			Property("name", gremlingo.T__.Select("c").Select("name")).
			Property("is_namespaced", gremlingo.T__.Select("c").Select("is_namespaced")).
			Property("namespace", gremlingo.T__.Select("c").Select("namespace")).
			Property("compromised", gremlingo.T__.Select("c").Select("compromised")).
			Property("critical", gremlingo.T__.Select("c").Select("critical"))
		return traversal
	}
}
