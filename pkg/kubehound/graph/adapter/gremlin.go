package adapter

import (
	"context"
	"encoding/json"
	"fmt"
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
