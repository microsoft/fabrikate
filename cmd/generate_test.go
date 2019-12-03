package cmd

import (
	"io/ioutil"
	"regexp"
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
	components, err := Generate("../testdata/generate-yaml", []string{"prod"}, false)

	expectedLengths := map[string]int{
		"prometheus-grafana": 125,
		"grafana":            8575,
		"prometheus":         21401,
	}

	assert.Nil(t, err)

	assert.Equal(t, 3, len(components))

	checkComponentLengthsAgainstExpected(t, components, expectedLengths)
}

func TestGenerateEnvYAML(t *testing.T) {
	_, err := Generate("../testdata/generate-yaml", []string{"common"}, false)

	assert.Nil(t, err)

	dat, _ := ioutil.ReadFile("../testdata/generate-yaml/generated/common/grafana.yaml")

	var re = regexp.MustCompile(`- access: proxy`)
	matches := re.FindAllStringIndex(string(dat), -1)

	assert.Equal(t, len(matches), 1);
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
