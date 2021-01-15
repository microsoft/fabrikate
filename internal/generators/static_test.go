package generators

import (
	"testing"

	"github.com/microsoft/fabrikate/internal/core"
	"github.com/stretchr/testify/assert"
)

func TestStaticGenerator_Generate(t *testing.T) {
	component := core.Component{
		Name:         "foo",
		Path:         "",
		PhysicalPath: "../../testdata/invaliddir",
	}

	generator := &StaticGenerator{}
	_, err := generator.Generate(&component)
	assert.NotNil(t, err)
}

func TestGetStaticComponentPath(t *testing.T) {
	component := core.Component{
		Name:          "kv-flexvol",
		ComponentType: "static",
		Method:        "http",
		Source:        "https://raw.githubusercontent.com/Azure/kubernetes-keyvault-flexvol/master/deployment/kv-flexvol-installer.yaml",
	}

	expectedComponentPath := "components/kv-flexvol"
	componentPath, _ := GetStaticManifestsPath(component)

	assert.Equal(t, expectedComponentPath, componentPath)
}
