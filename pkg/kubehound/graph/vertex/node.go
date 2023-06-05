package vertex

import (
	"github.com/DataDog/KubeHound/pkg/kubehound/models/graph"
	gremlin "github.com/apache/tinkerpop/gremlin-go/driver"
)

const (
	nodeLabel = "Node"
)

var _ Vertex = (*Node)(nil)

type Node struct {
	graph.Node
}

func (v Node) Label() string {
	return nodeLabel
}

func (v Node) Traversal() VertexTraversal {
	return func(source *gremlin.GraphTraversalSource, inserts []TraversalInput) *gremlin.GraphTraversal {
		g := source.GetGraphTraversal()

		for _, insert := range inserts {
			i := insert.(*graph.Node)
			g = g.AddV(v.Label()).
				Property("id", i.StoreId).
				Property("name", i.Name)
		}

		return g
	}
}
