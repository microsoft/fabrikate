package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerate(t *testing.T) {
	components, err := Generate("../test/fixtures/generate", "prod")

	assert.Nil(t, err)

	assert.Equal(t, 6, len(components))

	assert.Equal(t, "elasticsearch", components[3].Name)
	assert.Equal(t, 15298, len(components[3].Definition))

	assert.Equal(t, "fluentd", components[4].Name)
	assert.Equal(t, 21021, len(components[4].Definition))

	assert.Equal(t, "kibana", components[5].Name)
	assert.Equal(t, 1973, len(components[5].Definition))
}
