package core

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoad(t *testing.T) {
	config := ComponentConfig{
		Path: "../test/fixtures/load",
	}

	err := config.Load("test")
	assert.Nil(t, err)

	assert.Equal(t, "bar", config.Config["foo"])
	assert.Equal(t, "myapp", config.Namespace)
}

func TestMerge(t *testing.T) {
	currentConfig := NewComponentConfig("../test/fixtures/merge")
	err := currentConfig.Load("current")
	assert.Nil(t, err)

	newConfig := NewComponentConfig("../test/fixtures/merge")
	err = newConfig.Load("new")
	assert.Nil(t, err)

	err = currentConfig.Merge(newConfig)
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

func TestSet(t *testing.T) {
	config := NewComponentConfig("../test/fixtures/load")

	err := config.Load("test")
	assert.Nil(t, err)

	// can override a config value successfully
	config.SetConfig([]string{}, []string{"foo"}, "fee")
	assert.Equal(t, "fee", config.Config["foo"])

	// can create a value successfully
	config.SetConfig([]string{}, []string{"new"}, "value")
	assert.Equal(t, "value", config.Config["new"])

	// can override a subcomponent value successfully
	config.SetConfig([]string{"myapp"}, []string{"zoo"}, "zee")
	assert.Equal(t, "zee", config.Subcomponents["myapp"].Config["zoo"])

	// can create a new deeper subcomponent config level successfully
	config.SetConfig([]string{"myapp"}, []string{"data", "storageClass"}, "fast")
	dataMap := config.Subcomponents["myapp"].Config["data"].(map[string]interface{})
	assert.Equal(t, "fast", dataMap["storageClass"])
}

func TestWriteYAML(t *testing.T) {
	_ = os.Remove("../test/fixtures/write/config/test.yaml")

	config := ComponentConfig{
		Path:          "../test/fixtures/write",
		Serialization: "yaml",
		Config: map[string]interface{}{
			"foo": "bar",
		},
		Subcomponents: map[string]ComponentConfig{
			"myapp": ComponentConfig{
				Config: map[string]interface{}{
					"zoo": "zar",
				},
			},
		},
	}

	err := config.Write("test")
	assert.Nil(t, err)

	configContents, err := ioutil.ReadFile("../test/fixtures/write/config/test.yaml")
	assert.Nil(t, err)
	assert.Equal(t, 70, len(configContents))
}

func TestWriteJSON(t *testing.T) {
	_ = os.Remove("../test/fixtures/write/config/test.json")

	config := ComponentConfig{
		Path:          "../test/fixtures/write",
		Serialization: "json",
		Config: map[string]interface{}{
			"foo": "bar",
		},
		Subcomponents: map[string]ComponentConfig{
			"myapp": ComponentConfig{
				Config: map[string]interface{}{
					"zoo": "zar",
				},
			},
		},
	}

	err := config.Write("test")
	assert.Nil(t, err)

	configContents, err := ioutil.ReadFile("../test/fixtures/write/config/test.json")
	assert.Nil(t, err)
	assert.Equal(t, 132, len(configContents))
}
