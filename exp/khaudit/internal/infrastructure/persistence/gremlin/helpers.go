package gremlin

import "slices"

func stringArray(array []string) []any {
	// Sort and compact namespaces.
	slices.Sort(array)
	clean := slices.Compact(array)

	// Convert to raw slice.
	var out []any
	for _, s := range clean {
		out = append(out, s)
	}

	return out
}
