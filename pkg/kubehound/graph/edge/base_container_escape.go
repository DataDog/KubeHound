package edge

import (
	"context"
	"fmt"

	"github.com/DataDog/KubeHound/pkg/kubehound/graph/adapter"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/converter"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type BaseContainerEscape struct {
	BaseEdge
}

type containerEscapeGroup struct {
	Node      primitive.ObjectID `bson:"node_id" json:"node"`
	Container primitive.ObjectID `bson:"_id" json:"container"`
}

func containerEscapeProcessor(ctx context.Context, oic *converter.ObjectIDConverter, edgeLabel string, entry any, attributes map[string]any) (any, error) {
	typed, ok := entry.(*containerEscapeGroup)
	if !ok {
		return nil, fmt.Errorf("invalid type passed to processor: %T", entry)
	}

	return adapter.GremlinEdgeProcessor(ctx, oic, edgeLabel, typed.Container, typed.Node, attributes)
}
func (e *BaseContainerEscape) Traversal() types.EdgeTraversal {
	return adapter.DefaultEdgeTraversal()
}
