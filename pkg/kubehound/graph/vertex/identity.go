package vertex

import (
	"context"

	"github.com/DataDog/KubeHound/pkg/config"
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
	cfg *config.VertexBuilderConfig
}

func (v *Identity) Initialize(cfg *config.VertexBuilderConfig) error {
	v.cfg = cfg
	return nil
}

func (v *Identity) Label() string {
	return IdentityLabel
}

func (v *Identity) BatchSize() int {
	return BatchSizeDefault
}

func (v *Identity) Processor(ctx context.Context, entry any) (any, error) {
	return adapter.GremlinVertexProcessor[*graph.Identity](ctx, entry)
}

func (v *Identity) Traversal() types.VertexTraversal {
	return func(source *gremlingo.GraphTraversalSource, inserts []any) *gremlingo.GraphTraversal {
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
