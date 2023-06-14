package vertex

import (
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/graph"
	gremlin "github.com/apache/tinkerpop/gremlin-go/driver"
)

const (
	PodLabel = "Pod"
)

var _ Builder = (*Pod)(nil)

type Pod struct {
}

func (v Pod) Label() string {
	return PodLabel
}

func (v Pod) BatchSize() int {
	return DefaultBatchSize
}

func (v Pod) Traversal() Traversal {
	return func(source *gremlin.GraphTraversalSource, inserts []types.TraversalInput) *gremlin.GraphTraversal {
		g := source.GetGraphTraversal()
		for _, i := range inserts {
			data := i.(*graph.Pod)
			g = g.AddV(v.Label()).
				Property("store_id", data.StoreID).
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
