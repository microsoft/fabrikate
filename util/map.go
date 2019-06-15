package util

import (
	"strings"
)

// CopyMap takes in a map and returns a deep copy of it; not modifying the original at all
func CopyMap(m map[string]interface{}) map[string]interface{} {
	cp := make(map[string]interface{})
	for k, v := range m {
		vm, ok := v.(map[string]interface{})
		if ok {
			cp[k] = CopyMap(vm)
		} else {
			cp[k] = v
		}
	}

	return cp
}

// MergeMap takes in two maps and does an immutable merge of the two; applying the new onto the original.
// This is done immutably; Neither input maps are modified and a deep copy of the original is created, modified, and returned.
func MergeMap(originalMap map[string]interface{}, newValues map[string]interface{}) map[string]interface{} {
	// Deep copy the original input map
	copy := CopyMap(originalMap)

	for k, v := range newValues {
		// If the original and new values are maps, recursively set the original to a setMap of both values
		originalNestedMap, originalValueIsMap := copy[k].(map[string]interface{})
		newNestedMap, newValueIsMap := v.(map[string]interface{})

		if originalValueIsMap && newValueIsMap {
			copy[k] = MergeMap(originalNestedMap, newNestedMap)
		} else {
			copy[k] = v
		}
	}

	return copy
}

// FlattenMap will flatten a map of maps to a flat map with keys being seperated by `keySeperator`
// This is an immutable function, the original map is not modified and a new map is returned
func FlattenMap(nestedMap map[string]interface{}, keySeperator string, parentKeys []string) map[string]interface{} {
	flattened := make(map[string]interface{})

	for k, v := range nestedMap {
		nestedMap, valueIsMap := v.(map[string]interface{})
		if valueIsMap {
			flattenedNestedMap := FlattenMap(nestedMap, keySeperator, append(parentKeys, k))
			for nk, nv := range flattenedNestedMap {
				flattened[nk] = nv
			}
		} else {
			flattenedKey := strings.Join(append(parentKeys, k), keySeperator)
			flattened[flattenedKey] = v
		}
	}

	return flattened
}
