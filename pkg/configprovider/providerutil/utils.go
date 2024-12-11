package providerutil

import (
	"go.opentelemetry.io/collector/confmap"
)

// Removes the key(path passed in keys array) from sourceMap if same key path is present in MergeMap
func removeMatchingKeyFromSrcMap(srcMap map[string]interface{}, mergeMap map[string]interface{}, keys []string) map[string]interface{} {

	if len(keys) == 0 {
		return srcMap
	}

	currentKey := keys[0]
	if len(keys) == 1 { // Got Leaf key, if leaf key exists in both maps, remove it from source map
		_, existInSrc := srcMap[currentKey]
		_, existInMerge := mergeMap[currentKey]
		// If leaf key exists in both maps, remove from source map
		if existInSrc && existInMerge {
			delete(srcMap, currentKey)
		}
		return srcMap
	}
	// More levels to go, descend into child if current key present in both maps
	srcNestedMap, isCurrKeyInSrcMap := srcMap[currentKey].(map[string]interface{})
	mergeNestedMap, isCurrKeyInMergeMap := mergeMap[currentKey].(map[string]interface{})
	if isCurrKeyInSrcMap && isCurrKeyInMergeMap {
		removeMatchingKeyFromSrcMap(srcNestedMap, mergeNestedMap, keys[1:])
	}
	return srcMap
}

// Checks and prepares scrConf for replace behavior instead of map merge
// Hack for confmap.Conf.Merge method to replace specific fields instead of merging them
// Merge method merges field values from source and mergeConf, so by removing existing values from
// source map, we can achieve replace behavior
func PrepareForReplaceBehavior(srcConf *confmap.Conf, mergeConf *confmap.Conf) {
    	collectorFieldsPath := []string{"extensions", "sumologic", "collector_fields"}
   	keyPathsWithReplaceBehavior := [][]string{
		 collectorFieldsPath,
    	}
	for _, keyPath := range keyPathsWithReplaceBehavior {
		*srcConf = *confmap.NewFromStringMap(removeMatchingKeyFromSrcMap(srcConf.ToStringMap(), mergeConf.ToStringMap(), keyPath))
	}
}
