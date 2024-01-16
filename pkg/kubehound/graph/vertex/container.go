package vertex

import (
	"context"

	"github.com/DataDog/KubeHound/pkg/kubehound/graph/adapter"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/graph"
)

const (
	ContainerLabel = "Container"
)

var _ Builder = (*Container)(nil)

type Container struct {
	BaseVertex
}

func (v *Container) Label() string {
	return ContainerLabel
}

func (v *Container) Processor(ctx context.Context, entry any) (any, error) {
	return adapter.GremlinVertexProcessor[*graph.Container](ctx, entry)
}

func (v *Container) Traversal() types.VertexTraversal {
	return v.DefaultTraversal(v.Label())
}
