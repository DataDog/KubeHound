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

// containerEscapeTraversal expects a list of containerEscapeGroup serialized as mapstructure for injection into the graph.
// For each containerEscapeGroup, the traversal will: 1) find the container with matching storeID, 2) find the
// container vertex with matching storeID, and 3) add a CE_{edgeLabel} edge between the two vertices.
func containerEscapeTraversal(edgeLabel string) Traversal {
	return func(source *gremlin.GraphTraversalSource, inserts []types.TraversalInput) *gremlin.GraphTraversal {
		g := source.GetGraphTraversal().
			Inject(inserts).
			Unfold().As("ce").
			V().HasLabel(vertex.ContainerLabel).
			Where(P.Eq("ce")).
			By("storeID").
			By("container").
			AddE(edgeLabel).
			To(
				__.V().HasLabel(vertex.NodeLabel).
					Where(P.Eq("ce")).
					By("storeID").
					By("node")).
			Barrier().Limit(0)

		return g
	}
}
