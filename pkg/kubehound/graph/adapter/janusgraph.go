package adapter

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/DataDog/KubeHound/pkg/kubehound/graph/types"
)

// GremlinProcessor is the default processor implementation to process data container object returned from a
// store query and transform them into TraversalInput objects for JanusGraph/Gremlin.
func GremlinProcessor[T any](_ context.Context, entry types.DataContainer) (map[string]any, error) {
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

// StructToMap transforms a struct to a map to be consumed by a gremlin graph traversal implementation.
// TODO: review implementation... surely there's a better way?
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
