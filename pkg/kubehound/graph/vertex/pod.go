package vertex

import (
	"context"

	"github.com/DataDog/KubeHound/pkg/kubehound/graph/adapter"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/graph"
	gremlin "github.com/apache/tinkerpop/gremlin-go/v3/driver"
)

const (
	PodLabel = "Pod"
)

var _ Builder = (*Pod)(nil)

type Pod struct {
}

func (v Pod) Label() string {
	return PodLabel
}

func (v Pod) BatchSize() int {
	return BatchSizeDefault
}

func (v Pod) Processor(ctx context.Context, entry any) (any, error) {
	return adapter.GremlinInputProcessor[*graph.Pod](ctx, entry)
}

func (v Pod) Traversal() types.VertexTraversal {
	return func(source *gremlin.GraphTraversalSource, inserts []types.TraversalInput) *gremlin.GraphTraversal {
		g := source.GetGraphTraversal().
			Inject(inserts).
			Unfold().As("pods").
			AddV(v.Label()).As("podVtx").
			Property("class", v.Label()). // labels are not indexed - use a mirror property
			SideEffect(
				__.Select("pods").
					Unfold().As("kv").
					Select("podVtx").
					Property(
						__.Select("kv").By(Column.Keys),
						__.Select("kv").By(Column.Values)))

		return g
	}
}
