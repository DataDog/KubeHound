package vertex

import (
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/graph"

	gremlin "github.com/apache/tinkerpop/gremlin-go/v3/driver"
)

const (
	RoleLabel = "Role"
)

var _ Builder = (*Role)(nil)

type Role struct {
}

func (v Role) Label() string {
	return RoleLabel
}

func (v Role) BatchSize() int {
	return DefaultBatchSize
}

func (v Role) Traversal() Traversal {
	return func(source *gremlin.GraphTraversalSource, inserts []types.TraversalInput) *gremlin.GraphTraversal {
		g := source.GetGraphTraversal()
		for _, i := range inserts {
			data := i.(*graph.Role)
			g = g.AddV(v.Label()).
				Property("class", v.Label()). // labels are not indexed - use a mirror property
				Property("storeID", data.StoreID).
				Property("name", data.Name).
				Property("isNamespaced", data.IsNamespaced).
				Property("namespace", data.Namespace)
			for _, rule := range data.Rules {
				g = g.Property(gremlin.Cardinality.Set, "rules", rule)
			}
		}
		return g
	}
}
