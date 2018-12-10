package models

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseComponentFromJson(t *testing.T) {
	componentJson, err := ioutil.ReadFile("../test/fixtures/definition/infra/component.json")
	assert.Nil(t, err)

	component, err := ParseComponentFromJson(componentJson)

	assert.Nil(t, err)
	assert.Equal(t, component.Name, "infra")
	assert.Equal(t, component.Type, "component")
	assert.Equal(t, len(component.Subcomponents), 1)
	assert.Equal(t, component.Subcomponents[0].Name, "efk")
	assert.Equal(t, component.Subcomponents[0].Source, "https://github.com/Microsoft/marina-elasticsearch-fluentd-kibana")
}
