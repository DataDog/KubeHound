package vertex

import (
	"context"

	"github.com/DataDog/KubeHound/pkg/kubehound/graph/adapter"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/graph"
	gremlingo "github.com/apache/tinkerpop/gremlin-go/v3/driver"
)

const (
	IdentityLabel = "Identity"
)

var _ Builder = (*Identity)(nil)

type Identity struct {
}

func (v Identity) Label() string {
	return IdentityLabel
}

func (v Identity) BatchSize() int {
	return DefaultBatchSize
}

func (v Identity) Processor(ctx context.Context, entry any) (any, error) {
	return adapter.GremlinInputProcessor[*graph.Identity](ctx, entry)
}

func (v Identity) Traversal() Traversal {
	return func(source *gremlingo.GraphTraversalSource, inserts []types.TraversalInput) *gremlingo.GraphTraversal {
		g := source.GetGraphTraversal().
			Inject(inserts).
			Unfold().As("ids").
			AddV(v.Label()).As("idVtx").
			Property("class", v.Label()). // labels are not indexed - use a mirror property
			SideEffect(
				__.Select("ids").
					Unfold().As("kv").
					Select("idVtx").
					Property(
						__.Select("kv").By(Column.Keys),
						__.Select("kv").By(Column.Values)))

		return g
	}
}
