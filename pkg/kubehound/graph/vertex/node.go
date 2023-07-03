package vertex

import (
	"context"

	"github.com/DataDog/KubeHound/pkg/kubehound/graph/adapter"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/graph"
	gremlin "github.com/apache/tinkerpop/gremlin-go/v3/driver"
)

const (
	NodeLabel = "Node"
)

var _ Builder = (*Node)(nil)

type Node struct {
}

func (v Node) Label() string {
	return NodeLabel
}

func (v Node) BatchSize() int {
	return BatchSizeDefault
}

func (v Node) Processor(ctx context.Context, entry any) (any, error) {
	return adapter.GremlinInputProcessor[*graph.Node](ctx, entry)
}

func (v Node) Traversal() types.VertexTraversal {
	return func(source *gremlin.GraphTraversalSource, inserts []types.TraversalInput) *gremlin.GraphTraversal {
		g := source.GetGraphTraversal().
			Inject(inserts).
			Unfold().As("nodes").
			AddV(v.Label()).As("nodeVtx").
			Property("class", v.Label()). // labels are not indexed - use a mirror property
			SideEffect(
				__.Select("nodes").
					Unfold().As("kv").
					Select("nodeVtx").
					Property(
						__.Select("kv").By(Column.Keys),
						__.Select("kv").By(Column.Values)))

		return g
	}
}
