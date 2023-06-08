package vertex

import (
	"github.com/DataDog/KubeHound/pkg/kubehound/models/graph"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	"github.com/DataDog/KubeHound/pkg/utils"
	gremlin "github.com/apache/tinkerpop/gremlin-go/driver"
	gremlingo "github.com/apache/tinkerpop/gremlin-go/driver"
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
	return func(g *gremlin.GraphTraversalSource, inserts []TraversalInput) *gremlin.GraphTraversal {
		insertsConverted := utils.ConvertSliceAnyToTyped[graph.Volume, TraversalInput](inserts)
		log.I.Infof(" ============== INSERTS Volumes ====== %+v", insertsConverted)
		traversal := g.Inject(inserts).Unfold().As("c").
			AddV(v.Label()).
			Property("store_id", gremlingo.T__.Select("c").Select("store_id")).
			Property("name", gremlingo.T__.Select("c").Select("name")).
			Property("type", gremlingo.T__.Select("c").Select("type")).
			Property("path", gremlingo.T__.Select("c").Select("path"))
		return traversal
	}
}
