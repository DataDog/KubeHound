package vertex

import (
	"context"

	"github.com/DataDog/KubeHound/pkg/kubehound/graph/adapter"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/graph"

	gremlin "github.com/apache/tinkerpop/gremlin-go/v3/driver"
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

func (v *PermissionSet) Processor(ctx context.Context, entry any) (any, error) {
	return adapter.GremlinVertexProcessor[*graph.Role](ctx, entry)
}

func (v *PermissionSet) Traversal() types.VertexTraversal {
	return func(source *gremlin.GraphTraversalSource, inserts []any) *gremlin.GraphTraversal {
		g := source.GetGraphTraversal().
			Inject(inserts).
			Unfold().As("roles").
			AddV(v.Label()).As("roleVtx").
			Property("class", v.Label()). // labels are not indexed - use a mirror property
			SideEffect(
				__.Select("roles").
					Unfold().As("kv").
					Select("roleVtx").
					Property(
						__.Select("kv").By(Column.Keys),
						__.Select("kv").By(Column.Values)))

		return g
	}
}
