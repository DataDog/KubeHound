package vertex

import (
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/graph"
	gremlin "github.com/apache/tinkerpop/gremlin-go/v3/driver"
)

const (
	NodeLabel = "Node"
)

var _ Builder = (*Node)(nil)

type Node struct {
}

func (v Node) Label() string {
	return NodeLabel
}

func (v Node) BatchSize() int {
	return DefaultBatchSize
}

func (v Node) Traversal() Traversal {
	return func(source *gremlin.GraphTraversalSource, inserts []types.TraversalInput) *gremlin.GraphTraversal {
		g := source.GetGraphTraversal()
		for _, w := range inserts {
			data := w.(*graph.Node)
			g = g.AddV(v.Label()).
				Property("class", v.Label()). // labels are not indexed - use a mirror property
				Property("storeID", data.StoreID).
				Property("name", data.Name).
				Property("isNamespaced", data.IsNamespaced).
				Property("namespace", data.Namespace).
				Property("compromised", int(data.Compromised)).
				Property("critical", data.Critical)
		}
		return g
	}
}
