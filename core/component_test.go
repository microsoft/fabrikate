package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadComponent(t *testing.T) {
	component := Component{
		PhysicalPath: "../test/fixtures/definition/infra",
		LogicalPath:  "infra",
	}

	component, err := component.LoadComponent()
	assert.Nil(t, err)

	assert.Nil(t, err)
	assert.Equal(t, component.Name, "infra")
	assert.Equal(t, component.Type, "component")
	assert.Equal(t, len(component.Subcomponents), 1)
	assert.Equal(t, component.Subcomponents[0].Name, "efk")
	assert.Equal(t, component.Subcomponents[0].Source, "https://github.com/Microsoft/marina-elasticsearch-fluentd-kibana")
}

func TestLoadConfig(t *testing.T) {
	component := Component{
		PhysicalPath: "../test/fixtures/generate/infra",
		LogicalPath:  "infra",
	}

	component, err := component.LoadComponent()
	assert.Nil(t, err)

	err = component.LoadConfig("prod")

	assert.Nil(t, err)
}

func TestIteratingDefinition(t *testing.T) {
	callbackCount := 0
	results, err := IterateComponentTree("../test/fixtures/iterator", "", func(path string, component *Component) (err error) {
		callbackCount++
		return nil
	})

	assert.Nil(t, err)
	assert.Equal(t, 3, len(results))
	assert.Equal(t, callbackCount, len(results))

	assert.Equal(t, results[1].PhysicalPath, "../test/fixtures/iterator/infra")
	assert.Equal(t, results[1].LogicalPath, "infra")

	assert.Equal(t, results[2].PhysicalPath, "../test/fixtures/iterator/infra/components/efk")
	assert.Equal(t, results[2].LogicalPath, "infra/efk")
}
