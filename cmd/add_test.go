package cmd

import (
	"os"
	"testing"

	"github.com/microsoft/fabrikate/core"
	"github.com/stretchr/testify/assert"
)

func TestAdd(t *testing.T) {

	// This test changes the cwd. Must change back so any tests following don't break
	cwd, err := os.Getwd()
	assert.Nil(t, err)
	defer func() {
		_ = os.Chdir(cwd)
	}()

	err = os.Chdir("../test/fixtures/add")
	assert.Nil(t, err)

	_ = os.Remove("./component.yaml")

	componentComponent := core.Component{
		Name:          "cloud-native",
		Source:        "https://github.com/timfpark/fabrikate-cloud-native",
		Method:        "git",
		ComponentType: "component",
	}

	err = Add(componentComponent)
	assert.Nil(t, err)

	helmComponent := core.Component{
		Name:          "elasticsearch",
		Source:        "https://github.com/helm/charts",
		Method:        "git",
		Path:          "stable/elasticsearch",
		ComponentType: "helm",
	}

	err = Add(helmComponent)
	assert.Nil(t, err)
}
