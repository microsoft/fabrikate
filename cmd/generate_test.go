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
	components, err := Generate("../testdata/generate", []string{"prod-east", "prod"}, false)

	assert.Nil(t, err)

	expectedLengths := map[string]int{
		"elasticsearch":         14477,
		"elasticsearch-curator": 2390,
		"fluentd-elasticsearch": 20230,
		"kibana":                1590,
		"static":                188,
	}

	assert.Equal(t, 8, len(components))

	checkComponentLengthsAgainstExpected(t, components, expectedLengths)
}

func TestGenerateYAML(t *testing.T) {
	components, err := Generate("../testdata/generate-yaml", []string{"prod"}, false)

	expectedLengths := map[string]int{
		"prometheus-grafana": 125,
		"grafana":            8552,
		"prometheus":         21341,
	}

	assert.Nil(t, err)

	assert.Equal(t, 3, len(components))

	checkComponentLengthsAgainstExpected(t, components, expectedLengths)
}

func TestGenerateStaticRemoteYAML(t *testing.T) {
	components, err := Generate("../testdata/generate-remote-static", []string{"common"}, false)

	expectedLengths := map[string]int{
		"keyvault-flexvolume": 5,
		"keyvault-sub":        1372,
	}

	assert.Nil(t, err)
	assert.Equal(t, 2, len(components))

	checkComponentLengthsAgainstExpected(t, components, expectedLengths)
}

func TestGenerateWithHooks(t *testing.T) {
	_, err := Generate("../testdata/generate-hooks", []string{"prod"}, false)

	assert.Nil(t, err)
}
