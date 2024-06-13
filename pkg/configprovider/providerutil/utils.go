package providerutil

import "go.opentelemetry.io/collector/confmap"

// Remove specific keys from srcMap if same keys present in mergeMap under same path
func removeMatchingKeysFromSrcMap(srcMap map[string]interface{}, mergeMap map[string]interface{}) map[string]interface{} {
	srcMapVal := srcMap
	mergeMapVal := mergeMap
	for key, mergeValue := range mergeMapVal {
		srcValue, exists := srcMapVal[key]

		//if key not exists in merge map, skip the path
		if !exists {
			continue
		}

		if key == "collector_fields" {
			delete(srcMapVal, key)
		}

		mergeNestedMap, isMergeMap := mergeValue.(map[string]interface{})
		srcNestedMap, isSrcMap := srcValue.(map[string]interface{})

		if isMergeMap && isSrcMap {
			// Recursively handle nested maps
			removeMatchingKeysFromSrcMap(srcNestedMap, mergeNestedMap)
		}
	}
	return srcMap
}

// Checks and prepares scrConf for replace behavior instead of map merge
// Hack for confmap.Conf.Merge method to replace specific fields instead of merging them
// Merge method merges field values from source and mergeConf, so by removing existing values from
// source map, we can achieve replace behavior
func PrepareForReplaceBehavior(srcConf *confmap.Conf, mergeConf *confmap.Conf) {
	*srcConf = *confmap.NewFromStringMap(removeMatchingKeysFromSrcMap(srcConf.ToStringMap(), mergeConf.ToStringMap()))
}
