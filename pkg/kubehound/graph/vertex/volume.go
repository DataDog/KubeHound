package vertex

import (
	"github.com/DataDog/KubeHound/pkg/kubehound/models/graph"
	gremlin "github.com/apache/tinkerpop/gremlin-go/driver"
)

const (
	volumeLabel = "Volume"
)

var _ Vertex = (*Volume)(nil)

type Volume struct {
}

func (v Volume) Label() string {
	return volumeLabel
}

func (v Volume) Traversal() VertexTraversal {
	return func(source *gremlin.GraphTraversalSource, inserts []TraversalInput) *gremlin.GraphTraversal {
		g := source.GetGraphTraversal()

		for _, insert := range inserts {
			i := insert.(*graph.Volume)
			g = g.AddV(v.Label()).
				Property("id", i.StoreId).
				Property("name", i.Name).
				Property("type", i.Type).
				Property("path", i.Path)
		}

		return g
	}
}
