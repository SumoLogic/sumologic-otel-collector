package providerutil

import (
	"go.opentelemetry.io/collector/confmap"
	"strings"
)

func removeKeyFromSrcMap(srcMap map[string]interface{}, mergeMap map[string]interface{}, keys []string, index int) map[string]interface{} {

	if index == len(keys) { //out of index
		return srcMap
	}

	currentKey := keys[index]
	leafIndex := len(keys) - 1
	if index == leafIndex { // Got Leaf key, if leaf key exists in both maps, remove it from source map
		_, existInSrc := srcMap[currentKey]
		_, existInMerge := mergeMap[currentKey]
		// If leaf key exists in both maps, remove from source map
		if existInSrc && existInMerge {
			delete(srcMap, currentKey)
		}
	} else { // More levels to go, descend into child if current key present in both maps
		srcNestedMap, isCurrKeyInSrcMap := srcMap[currentKey].(map[string]interface{})
		mergeNestedMap, isCurrKeyInMergeMap := mergeMap[currentKey].(map[string]interface{})
		if isCurrKeyInSrcMap && isCurrKeyInMergeMap {
			removeKeyFromSrcMap(srcNestedMap, mergeNestedMap, keys, index+1)
		}
	}
	return srcMap
}

// Checks and prepares scrConf for replace behavior instead of map merge
// Hack for confmap.Conf.Merge method to replace specific fields instead of merging them
// Merge method merges field values from source and mergeConf, so by removing existing values from
// source map, we can achieve replace behavior
func PrepareForReplaceBehavior(srcConf *confmap.Conf, mergeConf *confmap.Conf) {
	keyPathsWithReplaceBehavior := [][]string{
		{"extensions", "sumologic", "collector_fields"},
	}
	for _, path := range keyPathsWithReplaceBehavior {
		pathKeys := strings.Split(path, "#")
		*srcConf = *confmap.NewFromStringMap(removeKeyFromSrcMap(srcConf.ToStringMap(), mergeConf.ToStringMap(), pathKeys, 0))
	}
}
