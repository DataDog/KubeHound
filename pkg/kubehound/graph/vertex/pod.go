package vertex

import (
	"github.com/DataDog/KubeHound/pkg/kubehound/models/graph"
	gremlin "github.com/apache/tinkerpop/gremlin-go/driver"
)

const (
	podLabel = "Pod"
)

var _ Builder = (*Pod)(nil)

type Pod struct {
}

func (v Pod) Label() string {
	return podLabel
}

func (v Pod) BatchSize() int {
	return DefaultBatchSize
}

func (v Pod) Traversal() VertexTraversal {
	return func(source *gremlin.GraphTraversalSource, inserts []TraversalInput) *gremlin.GraphTraversal {
		g := source.GetGraphTraversal()
		for _, i := range inserts {
			data := i.(*graph.Pod)
			g = g.AddV(v.Label()).
				Property("store_id", data.StoreId).
				Property("name", data.Name).
				Property("is_namespaced", data.IsNamespaced).
				Property("namespace", data.Namespace).
				Property("sharedProcessNamespace", data.SharedProcessNamespace).
				Property("serviceAccount", data.ServiceAccount).
				Property("node", data.Node).
				Property("compromised", int(data.Compromised)).
				Property("critical", data.Critical)
		}
		return g
	}
}
