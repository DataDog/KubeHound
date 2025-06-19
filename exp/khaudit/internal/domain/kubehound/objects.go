package kubehound

import (
	"encoding/json"
	"fmt"
)

// HexTuple represents a tuple in a graph database.
// https://github.com/ontola/hextuples
type HexTuple struct {
	Subject   string
	Predicate string
	Value     string
	DataType  string
	Language  string
	Graph     string
}

// MarshalJSON marshals the HexTuple to a JSON array.
func (t *HexTuple) MarshalJSON() ([]byte, error) {
	return json.Marshal([]string{t.Subject, t.Predicate, t.Value, t.DataType, t.Language, t.Graph})
}

// UnmarshalJSON unmarshals the HexTuple from a JSON array.
func (t *HexTuple) UnmarshalJSON(data []byte) error {
	var values []string
	// Decode as a JSON array.
	if err := json.Unmarshal(data, &values); err != nil {
		return err
	}

	// Check the number of values.
	if len(values) != 6 {
		return fmt.Errorf("expected 6 values, got %d", len(values))
	}

	// Unmarshal the values.
	t.Subject = values[0]
	t.Predicate = values[1]
	t.Value = values[2]
	t.DataType = values[3]
	t.Language = values[4]
	t.Graph = values[5]

	return nil
}

// AttackPath is a path of vertices and edges.
type AttackPath []HexTuple
