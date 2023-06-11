package envoy

import "sort"

func sortStringKeys[V any](sourceMap map[string]V) []string {
	keys := []string{}

	for key := range sourceMap {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	return keys
}
