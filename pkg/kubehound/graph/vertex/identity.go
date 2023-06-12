package vertex

import (
	"github.com/DataDog/KubeHound/pkg/kubehound/models/graph"
	gremlingo "github.com/apache/tinkerpop/gremlin-go/driver"
)

const (
	identityLabel = "Identity"
)

var _ Builder = (*Identity)(nil)

type Identity struct {
}

func (v Identity) Label() string {
	return identityLabel
}

func (v Identity) BatchSize() int {
	return DefaultBatchSize
}

func (v Identity) Traversal() VertexTraversal {
	return func(source *gremlingo.GraphTraversalSource, inserts []TraversalInput) *gremlingo.GraphTraversal {
		g := source.GetGraphTraversal()
		for _, i := range inserts {
			data := i.(*graph.Identity)
			g = g.AddV(v.Label()).
				Property("storeID", data.StoreID).
				Property("name", data.Name).
				Property("isNamespaced", data.IsNamespaced).
				Property("namespace", data.Namespace).
				Property("type", data.Type)
		}
		return g
	}
}
