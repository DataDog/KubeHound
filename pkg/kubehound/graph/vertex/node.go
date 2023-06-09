package vertex

import (
	"fmt"

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

func (v Node) Traversal() VertexTraversal {
	return func(source *gremlin.GraphTraversalSource, inserts []TraversalInput) *gremlin.GraphTraversal {
		g := source.GetGraphTraversal()
		for _, w := range inserts {
			data := w.(*graph.Node)
			fmt.Printf("WHAT THE FUCK IS THIS INSERTS DATA??????????????????? (cast) %+v\n", data)
			g = g.AddV(v.Label()).
				Property("storeId", data.StoreId).
				Property("name", data.Name).
				Property("is_namespaced", data.IsNamespaced).
				Property("namespace", data.Namespace).
				Property("compromised", int(data.Compromised)).
				Property("critical", data.Critical)
		}
		return g
	}
}
