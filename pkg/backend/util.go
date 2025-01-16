//go:build no_backend

package backend

func mergeMaps(currentMap map[interface{}]interface{}, newMap map[interface{}]interface{}) map[interface{}]interface{} {
	mergedMap := make(map[interface{}]interface{}, len(currentMap))
	for k, v := range currentMap {
		mergedMap[k] = v
	}

	for k, v := range newMap {
		if v, ok := v.(map[interface{}]interface{}); ok {
			if bv, ok := mergedMap[k]; ok {
				if bv, ok := bv.(map[interface{}]interface{}); ok {
					mergedMap[k] = mergeMaps(bv, v)

					continue
				}
			}
		}
		mergedMap[k] = v
	}

	return mergedMap
}
