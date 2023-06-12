package vertex

import (
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/graph"
	gremlin "github.com/apache/tinkerpop/gremlin-go/driver"
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

func (v Node) Traversal() Traversal {
	return func(source *gremlin.GraphTraversalSource, inserts []types.TraversalInput) *gremlin.GraphTraversal {
		g := source.GetGraphTraversal()
		for _, w := range inserts {
			data := w.(*graph.Node)
			g = g.AddV(v.Label()).
				Property("storeID", data.StoreID).
				Property("name", data.Name).
				Property("is_namespaced", data.IsNamespaced).
				Property("namespace", data.Namespace).
				Property("compromised", int(data.Compromised)).
				Property("critical", data.Critical)
		}
		return g
	}
}
