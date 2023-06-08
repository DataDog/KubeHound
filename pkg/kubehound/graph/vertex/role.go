package vertex

import (
	gremlin "github.com/apache/tinkerpop/gremlin-go/driver"
	gremlingo "github.com/apache/tinkerpop/gremlin-go/driver"
)

const (
	roleLabel = "Role"
)

var _ Builder = (*Role)(nil)

type Role struct {
}

func (v Role) Label() string {
	return roleLabel
}

func (v Role) BatchSize() int {
	return DefaultBatchSize
}

func (v Role) Traversal() VertexTraversal {
	return func(g *gremlin.GraphTraversalSource, inserts []TraversalInput) *gremlin.GraphTraversal {
		traversal := g.Inject(inserts).Unfold().As("c").
			AddV(v.Label()).
			Property("store_id", gremlingo.T__.Select("c").Select("store_id")).
			Property("name", gremlingo.T__.Select("c").Select("name")).
			Property("is_namespaced", gremlingo.T__.Select("c").Select("is_namespaced")).
			Property("namespace", gremlingo.T__.Select("c").Select("namespace"))
			// Property("rules", gremlingo.T__.Select("c").Select("rules")) // array of values
		return traversal
	}
}
