package vertex

import (
	"context"

	"github.com/DataDog/KubeHound/pkg/kubehound/graph/adapter"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/graph"
)

const (
	NodeLabel = "Node"
)

var _ Builder = (*Node)(nil)

type Node struct {
	BaseVertex
}

func (v *Node) Label() string {
	return NodeLabel
}

func (v *Node) Processor(ctx context.Context, entry any) (any, error) {
	return adapter.GremlinVertexProcessor[*graph.Node](ctx, entry)
}

func (v *Node) Traversal() types.VertexTraversal {
	return v.DefaultTraversal(v.Label())
}
