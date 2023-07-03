package adapter

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/DataDog/KubeHound/pkg/kubehound/models/converter"
	gremlin "github.com/apache/tinkerpop/gremlin-go/v3/driver"
)

// GremlinInputProcessor transform a graph model object to a map suitable for consumption by a gremllin traversal.
func GremlinInputProcessor[T any](_ context.Context, entry any) (map[string]any, error) {
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

func ConstructEdgeMerge(ctx context.Context, oic *converter.ObjectIdConverter, label string,
	sidOut string, sidIn string) (map[any]any, error) {

	vidIn, err := oic.GraphId(ctx, sidIn)
	if err != nil {
		return nil, err
	}

	vidOut, err := oic.GraphId(ctx, sidOut)
	if err != nil {
		return nil, err
	}

	processed := map[any]any{
		gremlin.T.Label:       label,
		gremlin.Direction.In:  vidIn,
		gremlin.Direction.Out: vidOut,
	}

	return processed, nil
}
