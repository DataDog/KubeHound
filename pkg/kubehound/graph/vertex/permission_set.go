package vertex

import (
	"context"

	"github.com/DataDog/KubeHound/pkg/kubehound/graph/adapter"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/graph"
)

const (
	PermissionSetLabel = "PermissionSet"
)

var _ Builder = (*PermissionSet)(nil)

type PermissionSet struct {
	BaseVertex
}

func (v *PermissionSet) Label() string {
	return PermissionSetLabel
}

func (v *PermissionSet) BatchSize() int {
	return v.cfg.BatchSizeSmall
}

func (v *PermissionSet) Processor(ctx context.Context, entry any) (any, error) {
	return adapter.GremlinVertexProcessor[*graph.PermissionSet](ctx, entry)
}

func (v *PermissionSet) Traversal() types.VertexTraversal {
	return v.DefaultTraversal(v.Label())
}
