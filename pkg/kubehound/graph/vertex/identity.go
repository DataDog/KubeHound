package vertex

import (
	"github.com/DataDog/KubeHound/pkg/kubehound/models/graph"
	gremlin "github.com/apache/tinkerpop/gremlin-go/driver"
)

const (
	identityLabel = "Identity"
)

var _ Vertex = (*Identity)(nil)

type Identity struct {
}

func (v Identity) Label() string {
	return identityLabel
}

func (v Identity) Traversal() VertexTraversal {
	return func(source *gremlin.GraphTraversalSource, inserts []TraversalInput) *gremlin.GraphTraversal {
		g := source.GetGraphTraversal()

		for _, insert := range inserts {
			i := insert.(*graph.Identity)
			g = g.AddV(v.Label()).
				Property("id", i.StoreId).
				Property("name", i.Name).
				Property("is_namespaced", i.IsNamespaced).
				Property("namespace", i.Namespace).
				Property("type", i.Type)
		}

		return g
	}
}
