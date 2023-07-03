package vertex

import (
	"context"

	"github.com/DataDog/KubeHound/pkg/kubehound/graph/adapter"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/graph"
	gremlingo "github.com/apache/tinkerpop/gremlin-go/v3/driver"
)

const (
	ContainerLabel = "Container"
)

var _ Builder = (*Container)(nil)

type Container struct {
}

func (v Container) Label() string {
	return ContainerLabel
}

func (v Container) BatchSize() int {
	return BatchSizeDefault / 2
}

func (v Container) Processor(ctx context.Context, entry any) (any, error) {
	return adapter.GremlinInputProcessor[*graph.Container](ctx, entry)
}

func (v Container) Traversal() types.VertexTraversal {
	return func(source *gremlingo.GraphTraversalSource, inserts []types.TraversalInput) *gremlingo.GraphTraversal {
		g := source.GetGraphTraversal().
			Inject(inserts).
			Unfold().As("containers").
			AddV(v.Label()).As("containerVtx").
			Property("class", v.Label()). // labels are not indexed - use a mirror property
			SideEffect(
				__.Select("containers").
					Unfold().As("kv").
					Select("containerVtx").
					Property(
						__.Select("kv").By(Column.Keys),
						__.Select("kv").By(Column.Values)))

		return g
	}
}
