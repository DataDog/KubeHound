package adapter

import (
	"context"
	"fmt"

	"github.com/DataDog/KubeHound/pkg/kubehound/graph/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/converter"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/store"
	gremlin "github.com/apache/tinkerpop/gremlin-go/v3/driver"
)

// Optional syntactic sugar.
var __ = gremlin.T__

// GremlinEdgeProcessor transforms the inputs into a map suitable for bulk edge insert using the MergeE API.
func GremlinEdgeProcessor(ctx context.Context, oic *converter.ObjectIDConverter, label string,
	out int64, in int64, attributes map[string]any) (map[any]any, error) {

	vidIn, err := oic.GraphID(ctx, store.Hex(in))
	if err != nil {
		return nil, fmt.Errorf("%s edge IN id convert: %w", label, err)
	}

	vidOut, err := oic.GraphID(ctx, store.Hex(out))
	if err != nil {
		return nil, fmt.Errorf("%s edge OUT id convert: %w", label, err)
	}

	processed := map[any]any{
		gremlin.T.Label:       label,
		gremlin.Direction.In:  vidIn,
		gremlin.Direction.Out: vidOut,
	}

	// Add any additional attributes to the edge.
	for k, v := range attributes {
		switch k {
		case string(gremlin.T.Label), string(gremlin.T.Id), string(gremlin.Direction.In), string(gremlin.Direction.Out):
			// Skip reserved keys.
			continue
		}

		// Add the attribute to the edge.
		processed[k] = v
	}

	return processed, nil
}

// DefaultEdgeTraversal returns the traversal to insert a set of edges from a map using the MergeE API.
func DefaultEdgeTraversal() types.EdgeTraversal {
	return func(source *gremlin.GraphTraversalSource, inserts []any) *gremlin.GraphTraversal {
		g := source.GetGraphTraversal().
			Inject(inserts).
			Unfold().As("em").
			MergeE(__.Select("em")).
			Barrier().Limit(0)

		return g
	}
}
