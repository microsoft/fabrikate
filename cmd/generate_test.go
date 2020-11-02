package cmd

import (
	"testing"

	"github.com/microsoft/fabrikate/core"
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
	components, err := Generate("../testdata/generate", []string{"prod-east", "prod"}, false, false)

	assert.Nil(t, err)

	expectedLengths := map[string]int{
		"jaeger": 				26916,
		"static":                188,
	}

	assert.Equal(t, 4, len(components))

	checkComponentLengthsAgainstExpected(t, components, expectedLengths)
}

func TestGenerateYAML(t *testing.T) {
	components, err := Generate("../testdata/generate-yaml", []string{"prod"}, false, false)

	expectedLengths := map[string]int{
		"prometheus-grafana": 125,
		"grafana":            8552,
		"prometheus":         28363,
	}

	assert.Nil(t, err)

	assert.Equal(t, 3, len(components))

	checkComponentLengthsAgainstExpected(t, components, expectedLengths)
}

func TestGenerateStaticRemoteYAML(t *testing.T) {
	components, err := Generate("../testdata/generate-remote-static", []string{"common"}, false, false)

	expectedLengths := map[string]int{
		"keyvault-flexvolume": 5,
		"keyvault-sub":        1372,
	}

	assert.Nil(t, err)
	assert.Equal(t, 2, len(components))

	checkComponentLengthsAgainstExpected(t, components, expectedLengths)
}

func TestGenerateWithHooks(t *testing.T) {
	_, err := Generate("../testdata/generate-hooks", []string{"prod"}, false, false)

	assert.Nil(t, err)
}

func TestGenerateDisabledSubcomponent(t *testing.T) {
	components, err := Generate("../testdata/generate-disabled", []string{"disabled"}, false, false)

	expectedLengths := map[string]int{
		"disabled-stack": 0,
	}

	assert.Nil(t, err)
	assert.Equal(t, 1, len(components))

	checkComponentLengthsAgainstExpected(t, components, expectedLengths)
}
