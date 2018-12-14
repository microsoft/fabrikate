package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerate(t *testing.T) {
	components, err := Generate("../test/fixtures/generate", "prod")

	assert.Nil(t, err)

	assert.Equal(t, 7, len(components))

	assert.Equal(t, "static", components[3].Name)
	assert.Equal(t, 188, len(components[3].Manifest))

	assert.Equal(t, "elasticsearch", components[4].Name)
	assert.Equal(t, 13606, len(components[4].Manifest))

	assert.Equal(t, "fluentd", components[5].Name)
	assert.Equal(t, 19489, len(components[5].Manifest))

	assert.Equal(t, "kibana", components[6].Name)
	assert.Equal(t, 1556, len(components[6].Manifest))
}
