package vertex

import (
	"context"

	"github.com/DataDog/KubeHound/pkg/kubehound/graph/adapter"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/graph"
)

const (
	PodLabel = "Pod"
)

var _ Builder = (*Pod)(nil)

type Pod struct {
	BaseVertex
}

func (v *Pod) Label() string {
	return PodLabel
}

func (v *Pod) Processor(ctx context.Context, entry any) (any, error) {
	return adapter.GremlinVertexProcessor[*graph.Pod](ctx, entry)
}

func (v *Pod) Traversal() types.VertexTraversal {
	return v.DefaultTraversal(v.Label())
}
