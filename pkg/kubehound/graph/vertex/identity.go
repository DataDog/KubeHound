package vertex

import (
	"context"

	"github.com/DataDog/KubeHound/pkg/kubehound/graph/adapter"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/graph"
)

const (
	IdentityLabel = "Identity"
)

var _ Builder = (*Identity)(nil)

type Identity struct {
	BaseVertex
}

func (v *Identity) Label() string {
	return IdentityLabel
}

func (v *Identity) Processor(ctx context.Context, entry any) (any, error) {
	return adapter.GremlinVertexProcessor[*graph.Identity](ctx, entry)
}

func (v *Identity) Traversal() types.VertexTraversal {
	return v.DefaultTraversal(v.Label())
}
