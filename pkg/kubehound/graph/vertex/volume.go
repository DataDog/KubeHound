package vertex

import (
	"context"

	"github.com/DataDog/KubeHound/pkg/kubehound/graph/adapter"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/graph"
)

const (
	VolumeLabel = "Volume"
)

var _ Builder = (*Volume)(nil)

type Volume struct {
	BaseVertex
}

func (v *Volume) Label() string {
	return VolumeLabel
}

func (v *Volume) Processor(ctx context.Context, entry any) (any, error) {
	return adapter.GremlinVertexProcessor[*graph.Volume](ctx, entry)
}

func (v *Volume) Traversal() types.VertexTraversal {
	return v.DefaultTraversal(v.Label())
}
