package utils

import (
	"fmt"

	"github.com/DataDog/KubeHound/pkg/telemetry/log"
)

// func AnySliceToStringSlice(in []any) []string {
// 	s := make([]string, len(in))
// 	for i, v := range in {
// 		s[i] = fmt.Sprint(v)
// 	}
// 	return s
// }
// func AnySliceToIntSlice(in []int64) []int64 {
// 	s := make([]int, len(in))
// 	for i, v := range in {
// 		s[i] = v
// 	}
// 	return s
// }

func ToSliceOfAny[Tin any](s []Tin) []string {
	result := make([]string, len(s))
	for i, v := range s {
		result[i] = fmt.Sprint(v)
	}
	return result
}

func ConvertSliceAnyToTyped[T any, Tin any](data []Tin) []T {
	converted := make([]T, len(data))
	for _, d := range converted {
		converted = append(converted, d)
	}
	return converted
}

func ConvertToSliceMapAny[T any](inserts []T) []map[string]any {
	toStore := make([]map[string]any, len(inserts))
	for _, currentStruct := range inserts {
		m, err := StructToMap(currentStruct)
		if err != nil {
			log.I.Errorf("Failed to convert struct to map for Nodes: %+v", err)
		}
		toStore = append(toStore, m)
	}
	return toStore
}
