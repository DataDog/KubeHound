package vertex

import (
	gremlin "github.com/apache/tinkerpop/gremlin-go/driver"
	gremlingo "github.com/apache/tinkerpop/gremlin-go/driver"
)

const (
	podLabel = "Pod"
)

var _ Builder = (*Pod)(nil)

type Pod struct {
}

func (v Pod) Label() string {
	return podLabel
}

func (v Pod) BatchSize() int {
	return DefaultBatchSize
}

func (v Pod) Traversal() VertexTraversal {
	return func(g *gremlin.GraphTraversalSource, inserts []TraversalInput) *gremlin.GraphTraversal {
		traversal := g.Inject(inserts).Unfold().As("c").
			AddV(v.Label()).
			Property("store_id", gremlingo.T__.Select("c").Select("store_id")).
			Property("name", gremlingo.T__.Select("c").Select("name")).
			Property("is_namespaced", gremlingo.T__.Select("c").Select("is_namespaced")).
			Property("namespace", gremlingo.T__.Select("c").Select("namespace")).
			Property("sharedProcessNamespace", gremlingo.T__.Select("c").Select("sharedProcessNamespace")).
			Property("serviceAccount", gremlingo.T__.Select("c").Select("serviceAccount")).
			Property("node", gremlingo.T__.Select("c").Select("node")).
			Property("compromised", gremlingo.T__.Select("c").Select("compromised")).
			Property("critical", gremlingo.T__.Select("c").Select("critical"))
		return traversal
	}
}
