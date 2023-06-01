package vertex

import (
	"github.com/DataDog/KubeHound/pkg/kubehound/models/graph"
	gremlin "github.com/apache/tinkerpop/gremlin-go/driver"
)

const (
	podLabel = "Pod"
)

var _ Vertex = (*Pod)(nil)

type Pod struct {
}

func (v Pod) Label() string {
	return podLabel
}

func (v Pod) Traversal() VertexTraversal {
	return func(source *gremlin.GraphTraversalSource, inserts []TraversalInput) *gremlin.GraphTraversal {
		g := source.GetGraphTraversal()

		for _, insert := range inserts {
			i := insert.(*graph.Pod)
			g = g.AddV(v.Label()).
				Property("id", i.StoreId).
				Property("name", i.Name).
				Property("is_namespaced", i.IsNamespaced).
				Property("namespace", i.Namespace).
				Property("shared_process_namespace", i.SharedProcessNamespace).
				Property("service_account", i.ServiceAccount).
				Property("node", i.Node).
				Property("compromised", i.Compromised).
				Property("critical", i.Critical)
		}

		return g
	}
}
