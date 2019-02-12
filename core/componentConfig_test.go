package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMergeConfig(t *testing.T) {
	currentConfig := ComponentConfig{
		Config: map[string]interface{}{},
		Subcomponents: map[string]ComponentConfig{
			"jaeger": ComponentConfig{
				Config: map[string]interface{}{
					"provisionDataStore": map[string]interface{}{
						"elasticsearch": false,
						"mixed":         1,
						"slice": []interface{}{
							"astring",
							1,
							1.0,
							map[string]interface{}{
								"foo": 1,
								"bar": []interface{}{
									"bstring",
									1,
								},
							},
						},
					},
				},
			},
		},
	}

	newConfig := ComponentConfig{
		Config: map[string]interface{}{},
		Subcomponents: map[string]ComponentConfig{
			"jaeger": ComponentConfig{
				Config: map[string]interface{}{
					"provisionDataStore": map[string]interface{}{
						"cassandra":     false,
						"elasticsearch": true,
						"mixed":         "2",
					},
				},
			},
		},
	}

	err := currentConfig.Merge(newConfig)
	assert.Nil(t, err)

	jaegerSubcomponent := currentConfig.Subcomponents["jaeger"]
	provisionConfig := jaegerSubcomponent.Config["provisionDataStore"].(map[string]interface{})

	cassandraValue := provisionConfig["cassandra"].(bool)
	assert.Equal(t, false, cassandraValue)

	elasticsearchValue := provisionConfig["elasticsearch"].(bool)
	assert.Equal(t, false, elasticsearchValue)

	mixed := provisionConfig["mixed"].(int)
	assert.Equal(t, 1, mixed)

	sliceValues := provisionConfig["slice"].([]interface{})
	assert.Equal(t, 1, sliceValues[1].(int))
}
