package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFlattenMap(t *testing.T) {
	nestedMap := map[string]interface{}{
		"foo": "bar",
		"im": map[string]interface{}{
			"a": map[string]interface{}{
				"really": map[string]interface{}{
					"nested": "map",
				},
				"list": []int{1, 2, 3},
			}},
	}

	flattenedWithDots := FlattenMap(nestedMap, ".", []string{})
	assert.EqualValues(t, map[string]interface{}{
		"foo":                "bar",
		"im.a.really.nested": "map",
		"im.a.list":          []int{1, 2, 3},
	}, flattenedWithDots)

	flattenedWithDashes := FlattenMap(nestedMap, "-", []string{})
	assert.EqualValues(t, map[string]interface{}{
		"foo":                "bar",
		"im-a-really-nested": "map",
		"im-a-list":          []int{1, 2, 3},
	}, flattenedWithDashes)
}
