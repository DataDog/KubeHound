package vertex

import (
	"github.com/DataDog/KubeHound/pkg/kubehound/models/graph"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	"github.com/DataDog/KubeHound/pkg/utils"
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
		insertsConverted := utils.ConvertSliceAnyToTyped[graph.Pod, TraversalInput](inserts)
		toStore := utils.ConvertToSliceMapAny(insertsConverted)
		log.I.Infof(" ============== INSERTS Pods ====== %+v", insertsConverted)
		log.I.Infof(" ============== toStore Pods ====== %+v", toStore)
		g := source.GetGraphTraversal()
		for _, i := range toStore {
			g = g.AddV(v.Label()).
				Property("storeId", i["storeId"]).
				Property("name", i["name"]).
				Property("isNamespaced", i["isNamespaced"]).
				Property("namespace", i["namespace"]).
				Property("type", i["type"])
		}
		return g
		// return g.Inject(toStore).Unfold().As("c").
		// 	AddV(v.Label()).
		// 	Property("store_id", gremlingo.T__.Select("c").Select("store_id")).
		// 	Property("name", gremlingo.T__.Select("c").Select("name")).
		// 	Property("is_namespaced", gremlingo.T__.Select("c").Select("is_namespaced")).
		// 	Property("namespace", gremlingo.T__.Select("c").Select("namespace")).
		// 	Property("sharedProcessNamespace", gremlingo.T__.Select("c").Select("sharedProcessNamespace")).
		// 	Property("serviceAccount", gremlingo.T__.Select("c").Select("serviceAccount")).
		// 	Property("node", gremlingo.T__.Select("c").Select("node")).
		// 	Property("compromised", gremlingo.T__.Select("c").Select("compromised")).
		// 	Property("critical", gremlingo.T__.Select("c").Select("critical"))
	}
}
