package cmd

import (
	"testing"

	"github.com/Microsoft/fabrikate/core"
	"github.com/stretchr/testify/assert"
)

func checkComponentLengthsAgainstExpected(t *testing.T, components []core.Component, expectedLengths map[string]int) {
	for _, component := range components {
		if expectedLength, ok := expectedLengths[component.Name]; ok {
			assert.True(t, ok)

			assert.Equal(t, expectedLength, len(component.Manifest))
		}
	}
}

func TestGenerateJSON(t *testing.T) {
	components, err := Generate("../test/fixtures/generate", []string{"prod-east", "prod"})

	assert.Nil(t, err)

	expectedLengths := map[string]int{
		"elasticsearch":         14495,
		"elasticsearch-curator": 2394,
		"fluentd-elasticsearch": 20203,
		"kibana":                1595,
		"static":                188,
	}

	assert.Equal(t, 8, len(components))

	checkComponentLengthsAgainstExpected(t, components, expectedLengths)
}

func TestGenerateYAML(t *testing.T) {
	components, err := Generate("../test/fixtures/generate-yaml", []string{"prod"})

	expectedLengths := map[string]int{
		"prometheus-grafana": 125,
		"grafana":            8575,
		"prometheus":         21401,
	}

	assert.Nil(t, err)

	assert.Equal(t, 3, len(components))

	checkComponentLengthsAgainstExpected(t, components, expectedLengths)
}

func TestGenerateWithHooks(t *testing.T) {
	_, err := Generate("../test/fixtures/generate-hooks", []string{"prod"})

	assert.Nil(t, err)
}
