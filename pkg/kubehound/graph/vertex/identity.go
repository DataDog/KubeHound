package vertex

import (
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/graph"
	gremlingo "github.com/apache/tinkerpop/gremlin-go/v3/driver"
)

const (
	IdentityLabel = "Identity"
)

var _ Builder = (*Identity)(nil)

type Identity struct {
}

func (v Identity) Label() string {
	return IdentityLabel
}

func (v Identity) BatchSize() int {
	return DefaultBatchSize
}

func (v Identity) Traversal() Traversal {
	return func(source *gremlingo.GraphTraversalSource, inserts []types.TraversalInput) *gremlingo.GraphTraversal {
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
