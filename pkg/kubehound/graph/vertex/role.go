package vertex

import (
	"github.com/DataDog/KubeHound/pkg/kubehound/models/graph"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	"github.com/DataDog/KubeHound/pkg/utils"
	gremlin "github.com/apache/tinkerpop/gremlin-go/driver"
)

const (
	roleLabel = "Role"
)

var _ Builder = (*Role)(nil)

type Role struct {
}

func (v Role) Label() string {
	return roleLabel
}

func (v Role) BatchSize() int {
	return DefaultBatchSize
}

func (v Role) Traversal() VertexTraversal {
	return func(source *gremlin.GraphTraversalSource, inserts []TraversalInput) *gremlin.GraphTraversal {
		insertsConverted := utils.ConvertSliceAnyToTyped[graph.Role, TraversalInput](inserts)
		toStore := utils.ConvertToSliceMapAny(insertsConverted)
		log.I.Infof(" ============== INSERTS Role ====== %+v", insertsConverted)
		log.I.Infof(" ============== toStore Role ====== %+v", toStore)
		g := source.GetGraphTraversal()
		for _, i := range toStore {
			g = g.AddV(v.Label()).
				Property("storeId", i["storeId"]).
				Property("name", i["name"]).
				Property("is_namespaced", i["is_namespaced"]).
				Property("namespace", i["namespace"])
		}
		return g
		// return g.Inject(toStore).Unfold().As("c").
		// 	AddV(v.Label()).
		// 	Property("store_id", gremlingo.T__.Select("c").Select("store_id")).
		// 	Property("name", gremlingo.T__.Select("c").Select("name")).
		// 	Property("is_namespaced", gremlingo.T__.Select("c").Select("is_namespaced")).
		// 	Property("namespace", gremlingo.T__.Select("c").Select("namespace"))
		// // Property("rules", gremlingo.T__.Select("c").Select("rules")) // array of values
	}
}
