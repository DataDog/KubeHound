package vertex

import (
	"github.com/DataDog/KubeHound/pkg/kubehound/models/graph"
	gremlin "github.com/apache/tinkerpop/gremlin-go/driver"
)

const (
	roleLabel = "Role"
)

var _ Vertex = (*Role)(nil)

type Role struct {
}

func (v Role) Label() string {
	return roleLabel
}

func (v Role) Traversal() VertexTraversal {
	return func(source *gremlin.GraphTraversalSource, inserts []TraversalInput) *gremlin.GraphTraversal {
		g := source.GetGraphTraversal()

		for _, insert := range inserts {
			i := insert.(*graph.Role)
			g = g.AddV(v.Label()).
				Property("id", i.StoreId).
				Property("name", i.Name).
				Property("is_namespaced", i.IsNamespaced).
				Property("namespace", i.Namespace).
				Property("rules", i.Rules)
		}

		return g
	}
}
