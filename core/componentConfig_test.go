package core

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoad(t *testing.T) {
	config := ComponentConfig{
		Path: "../testdata/load",
	}

	err := config.Load("test")
	assert.Nil(t, err)

	assert.Equal(t, "bar", config.Config["foo"])
	assert.Equal(t, "myapp", config.Namespace)
}

func TestFailedYAMLLoad(t *testing.T) {
	config := ComponentConfig{
		Path: "../testdata/badyamlconfig",
	}

	err := config.Load("common")
	assert.NotNilNil(t, err)
}

func TestFailedJSONLoad(t *testing.T) {
	config := ComponentConfig{
		Path: "../testdata/badjsonconfig",
	}

	err := config.Load("common")
	assert.NotNilNil(t, err)
}

func TestMerge(t *testing.T) {
	currentConfig := NewComponentConfig("../testdata/merge")
	err := currentConfig.Load("current")
	assert.Nil(t, err)

	newConfig := NewComponentConfig("../testdata/merge")
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
	config := NewComponentConfig("../testdata/load")

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
	_ = os.Remove("../testdata/write/config/test.yaml")

	config := ComponentConfig{
		Path:          "../testdata/write",
		Serialization: "yaml",
		Config: map[string]interface{}{
			"foo": "bar",
		},
		Subcomponents: map[string]ComponentConfig{
			"myapp": {
				Config: map[string]interface{}{
					"zoo": "zar",
				},
			},
		},
	}

	err := config.Write("test")
	assert.Nil(t, err)

	configContents, err := ioutil.ReadFile("../testdata/write/config/test.yaml")
	assert.Nil(t, err)
	assert.Equal(t, 70, len(configContents))
}

func TestWriteJSON(t *testing.T) {
	_ = os.Remove("../testdata/write/config/test.json")

	config := ComponentConfig{
		Path:          "../testdata/write",
		Serialization: "json",
		Config: map[string]interface{}{
			"foo": "bar",
		},
		Subcomponents: map[string]ComponentConfig{
			"myapp": {
				Config: map[string]interface{}{
					"zoo": "zar",
				},
			},
		},
	}

	err := config.Write("test")
	assert.Nil(t, err)

	configContents, err := ioutil.ReadFile("../testdata/write/config/test.json")
	assert.Nil(t, err)
	assert.Equal(t, 132, len(configContents))
}
