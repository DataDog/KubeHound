package vertex

import (
	"github.com/DataDog/KubeHound/pkg/kubehound/models/graph"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	"github.com/DataDog/KubeHound/pkg/utils"
	gremlingo "github.com/apache/tinkerpop/gremlin-go/driver"
)

const (
	identityLabel = "Identity"
)

var _ Builder = (*Identity)(nil)

type Identity struct {
}

func (v Identity) Label() string {
	return identityLabel
}

func (v Identity) BatchSize() int {
	return DefaultBatchSize
}

func (v Identity) Traversal() VertexTraversal {
	return func(source *gremlingo.GraphTraversalSource, inserts []TraversalInput) *gremlingo.GraphTraversal {
		insertsConverted := utils.ConvertSliceAnyToTyped[graph.Identity, TraversalInput](inserts)
		toStore := utils.ConvertToSliceMapAny(insertsConverted)
		log.I.Infof(" ============== INSERTS Identity ====== %+v", insertsConverted)
		log.I.Infof(" ============== toStore Identity ====== %+v", toStore)
		g := source.GetGraphTraversal()
		for _, i := range toStore {
			// i := insert.(map[string]any)
			// command := utils.ToSliceOfAny(i["command"].([]any))
			// args := utils.ToSliceOfAny(i["args"].([]any))
			// capabilities := utils.ToSliceOfAny(i["capabilities"].([]any))

			g = g.AddV(v.Label()).
				Property("storeId", i["storeId"]).
				Property("name", i["name"]).
				Property("isNamespaced", i["isNamespaced"]).
				Property("namespace", i["namespace"]).
				Property("type", i["type"])
		}
		return g
		// return g.Inject(toStore).Unfold().As("c").AddV(v.Label()).
		// 	Property("storeId", gremlingo.T__.Select("c").Select("store_id")).
		// 	Property("name", gremlingo.T__.Select("c").Select("name")).
		// 	Property("isNamespaced", gremlingo.T__.Select("c").Select("is_namespaced")).
		// 	Property("namespace", gremlingo.T__.Select("c").Select("namespace")).
		// 	Property("type", gremlingo.T__.Select("c").Select("type"))
	}
}
