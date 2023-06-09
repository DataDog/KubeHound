package vertex

import (
	"github.com/DataDog/KubeHound/pkg/kubehound/models/graph"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	"github.com/DataDog/KubeHound/pkg/utils"
	gremlin "github.com/apache/tinkerpop/gremlin-go/driver"
)

const (
	volumeLabel = "Volume"
)

var _ Builder = (*Volume)(nil)

type Volume struct {
}

func (v Volume) Label() string {
	return volumeLabel
}

func (v Volume) BatchSize() int {
	return DefaultBatchSize
}

func (v Volume) Traversal() VertexTraversal {
	return func(source *gremlin.GraphTraversalSource, inserts []TraversalInput) *gremlin.GraphTraversal {
		insertsConverted := utils.ConvertSliceAnyToTyped[graph.Volume, TraversalInput](inserts)
		toStore := utils.ConvertToSliceMapAny(insertsConverted)
		log.I.Infof(" ============== INSERTS Volume ====== %+v", insertsConverted)
		log.I.Infof(" ============== toStore Volume ====== %+v", toStore)
		g := source.GetGraphTraversal()
		for _, i := range toStore {
			g = g.AddV(v.Label()).
				Property("storeId", i["storeId"]).
				Property("name", i["name"]).
				Property("type", i["type"]).
				Property("path", i["path"])
		}
		return g
		// return g.Inject(inserts).Unfold().As("c").
		// 	AddV(v.Label()).
		// 	Property("store_id", gremlingo.T__.Select("c").Select("store_id")).
		// 	Property("name", gremlingo.T__.Select("c").Select("name")).
		// 	Property("type", gremlingo.T__.Select("c").Select("type")).
		// 	Property("path", gremlingo.T__.Select("c").Select("path"))
	}
}
