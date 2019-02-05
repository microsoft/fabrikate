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
}
