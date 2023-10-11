package adapter

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/DataDog/KubeHound/pkg/kubehound/graph/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/converter"
	gremlin "github.com/apache/tinkerpop/gremlin-go/v3/driver"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Optional syntactic sugar.
var __ = gremlin.T__

// GremlinVertexProcessor transform a graph model object to a map suitable for consumption by a gremllin traversal.
func GremlinVertexProcessor[T any](_ context.Context, entry any) (map[string]any, error) {
	typed, ok := entry.(T)
	if !ok {
		return nil, fmt.Errorf("invalid type passed to processor: %T", entry)
	}

	processed, err := structToMap(typed)
	if err != nil {
		return nil, err
	}

	return processed, nil
}

// structToMap creates a map from a simple input struct.
func structToMap(in any) (map[string]any, error) {
	var res map[string]any

	tmp, err := json.Marshal(in)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(tmp, &res)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// GremlinEdgeProcessor transforms the inputs into a map suitable for bulk edge insert using the MergeE API.
func GremlinEdgeProcessor(ctx context.Context, oic *converter.ObjectIDConverter, label string,
	out primitive.ObjectID, in primitive.ObjectID) (map[any]any, error) {

	vidIn, err := oic.GraphID(ctx, in.Hex())
	if err != nil {
		return nil, fmt.Errorf("%s edge IN id convert: %w", label, err)
	}

	vidOut, err := oic.GraphID(ctx, out.Hex())
	if err != nil {
		return nil, fmt.Errorf("%s edge OUT id convert: %w", label, err)
	}

	processed := map[any]any{
		gremlin.T.Label:       label,
		gremlin.Direction.In:  vidIn,
		gremlin.Direction.Out: vidOut,
	}

	return processed, nil
}

// DefaultEdgeTraversal returns the traversal to insert a set of edges from a map using the MergeE API.
func DefaultEdgeTraversal() types.EdgeTraversal {
	return func(source *gremlin.GraphTraversalSource, inserts []any) *gremlin.GraphTraversal {
		g := source.GetGraphTraversal().
			//nolint:asasalint
			Inject(inserts).
			Unfold().As("em").
			MergeE(__.Select("em")).
			Barrier().Limit(0)

		return g
	}
}
