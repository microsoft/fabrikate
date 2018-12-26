package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerate(t *testing.T) {
	components, err := Generate("../test/fixtures/generate", "prod")

	assert.Nil(t, err)

	assert.Equal(t, 7, len(components))

	assert.Equal(t, "static", components[6].Name)
	assert.Equal(t, 188, len(components[6].Manifest))

	assert.Equal(t, "elasticsearch", components[3].Name)
	assert.Equal(t, 13638, len(components[3].Manifest))

	assert.Equal(t, "fluentd-elasticsearch", components[4].Name)
	assert.Equal(t, 20173, len(components[4].Manifest))

	assert.Equal(t, "kibana", components[5].Name)
	assert.Equal(t, 1595, len(components[5].Manifest))
}
