package vertex

import (
	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/types"
	gremlingo "github.com/apache/tinkerpop/gremlin-go/v3/driver"
)

type BaseVertex struct {
	cfg     *config.VertexBuilderConfig
	runtime *config.DynamicConfig
}

func (v *BaseVertex) Initialize(cfg *config.KubehoundConfig) error {
	v.cfg = &cfg.Builder.Vertex
	v.runtime = &cfg.Dynamic

	return nil
}

func (v *BaseVertex) BatchSize() int {
	return v.cfg.BatchSize
}

func (v *BaseVertex) DefaultTraversal(label string) types.VertexTraversal {
	return func(source *gremlingo.GraphTraversalSource, inserts []any) *gremlingo.GraphTraversal {
		g := source.GetGraphTraversal().
			//nolint:asasalint // required due to constraints in the gremlin API
			Inject(inserts).
			Unfold().As("entities").
			AddV(label).As("vtx").
			Property("class", label). // labels are not indexed - use a mirror property
			Property("cluster", v.runtime.Cluster).
			Property("runID", v.runtime.RunID).
			SideEffect(
				__.Select("entities").
					Unfold().As("kv").
					Select("vtx").
					Property(
						__.Select("kv").By(Column.Keys),
						__.Select("kv").By(Column.Values)))

		return g
	}
}
