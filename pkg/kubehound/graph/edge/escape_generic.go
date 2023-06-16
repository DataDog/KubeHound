package edge

import (
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/vertex"
	gremlin "github.com/apache/tinkerpop/gremlin-go/v3/driver"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type containerEscapeGroup struct {
	Node      primitive.ObjectID `bson:"node_id" json:"node"`
	Container primitive.ObjectID `bson:"_id" json:"container"`
}

func containerEscapeTraversal(edgeLabel string) Traversal {
	return func(source *gremlin.GraphTraversalSource, inserts []types.TraversalInput) *gremlin.GraphTraversal {
		g := source.GetGraphTraversal()

		for _, i := range inserts {
			ml := i.(*containerEscapeGroup)

			g = g.V().
				Has("class", vertex.ContainerLabel).
				Has("storeID", ml.Container.Hex()).
				As("container").
				V().
				Has("class", vertex.NodeLabel).
				Has("storeID", ml.Node.Hex()).
				As("node").
				AddE(edgeLabel).
				From("container").
				To("node")
		}

		return g
	}
}
