package utils

import "encoding/json"

// structToMap transforms a struct to a map to be consumed by a mongoDB AsyncWriter implementation.
// TODO: review implementation... surely there's a better way?
func StructToMap(in any) (map[string]any, error) {
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
