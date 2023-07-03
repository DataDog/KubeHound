package edge

import (
	"context"
	"fmt"

	"github.com/DataDog/KubeHound/pkg/kubehound/graph/adapter"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/converter"
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
			MergeE(__.Select("ce")).
			Barrier().Limit(0)

		return g
	}
}

func containerEscapeProcessor(ctx context.Context, oic *converter.ObjectIdConverter, edgeLabel string, entry any) (any, error) {
	typed, ok := entry.(*containerEscapeGroup)
	if !ok {
		return nil, fmt.Errorf("invalid type passed to processor: %T", entry)
	}

	return adapter.ConstructEdgeMerge(ctx, oic, edgeLabel, typed.Container.Hex(), typed.Node.Hex())
}
