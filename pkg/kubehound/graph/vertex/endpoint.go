package vertex

import (
	"context"

	"github.com/DataDog/KubeHound/pkg/kubehound/graph/adapter"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/graph"
	gremlin "github.com/apache/tinkerpop/gremlin-go/v3/driver"
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
	return func(source *gremlin.GraphTraversalSource, inserts []any) *gremlin.GraphTraversal {
		g := source.GetGraphTraversal().
			//nolint:asasalint
			Inject(inserts).
			Unfold().As("endpoints").
			AddV(v.Label()).As("epVtx").
			Property("class", v.Label()). // labels are not indexed - use a mirror property
			SideEffect(
				__.Select("endpoints").
					Unfold().As("kv").
					Select("epVtx").
					Property(
						__.Select("kv").By(Column.Keys),
						__.Select("kv").By(Column.Values)))

		return g
	}
}
