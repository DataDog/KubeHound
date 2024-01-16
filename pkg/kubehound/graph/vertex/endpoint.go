package vertex

import (
	"context"

	"github.com/DataDog/KubeHound/pkg/kubehound/graph/adapter"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/graph"
)

const (
	EndpointLabel = "Endpoint"
)

var _ Builder = (*Endpoint)(nil)

type Endpoint struct {
	BaseVertex
}

func (v *Endpoint) Label() string {
	return EndpointLabel
}

func (v *Endpoint) Processor(ctx context.Context, entry any) (any, error) {
	return adapter.GremlinVertexProcessor[*graph.Endpoint](ctx, entry)
}

func (v *Endpoint) Traversal() types.VertexTraversal {
	return v.DefaultTraversal(v.Label())
}
