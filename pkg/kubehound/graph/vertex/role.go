package vertex

import (
	"context"

	"github.com/DataDog/KubeHound/pkg/kubehound/graph/adapter"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/graph"

	gremlin "github.com/apache/tinkerpop/gremlin-go/v3/driver"
)

const (
	RoleLabel = "Role"
)

var _ Builder = (*Role)(nil)

type Role struct {
}

func (v Role) Label() string {
	return RoleLabel
}

func (v Role) BatchSize() int {
	return DefaultBatchSize
}

func (v Role) Processor(ctx context.Context, entry any) (any, error) {
	return adapter.GremlinInputProcessor[*graph.Role](ctx, entry)
}

func (v Role) Traversal() Traversal {
	return func(source *gremlin.GraphTraversalSource, inserts []types.TraversalInput) *gremlin.GraphTraversal {
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
