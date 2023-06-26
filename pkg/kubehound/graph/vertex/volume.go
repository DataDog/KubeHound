package vertex

import (
	"context"

	"github.com/DataDog/KubeHound/pkg/kubehound/graph/adapter"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/graph"
	gremlin "github.com/apache/tinkerpop/gremlin-go/v3/driver"
)

const (
	VolumeLabel = "Volume"
)

var _ Builder = (*Volume)(nil)

type Volume struct {
}

func (v Volume) Label() string {
	return VolumeLabel
}

func (v Volume) BatchSize() int {
	return DefaultBatchSize
}

func (v Volume) Processor(ctx context.Context, entry any) (any, error) {
	return adapter.GremlinInputProcessor[*graph.Volume](ctx, entry)
}

func (v Volume) Traversal() Traversal {
	return func(source *gremlin.GraphTraversalSource, inserts []types.TraversalInput) *gremlin.GraphTraversal {
		g := source.GetGraphTraversal().
			Inject(inserts).
			Unfold().As("volumes").
			AddV(v.Label()).As("volVtx").
			Property("class", v.Label()). // labels are not indexed - use a mirror property
			SideEffect(
				__.Select("volumes").
					Unfold().As("kv").
					Select("volVtx").
					Property(
						__.Select("kv").By(Column.Keys),
						__.Select("kv").By(Column.Values))).
			Barrier().Limit(0)

		return g
	}
}
