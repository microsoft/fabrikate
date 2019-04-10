package cmd

import (
	"os"
	"testing"

	"github.com/Microsoft/fabrikate/core"
	"github.com/stretchr/testify/assert"
)

func TestAdd(t *testing.T) {
	err := os.Chdir("../test/fixtures/add")
	assert.Nil(t, err)

	_ = os.Remove("./component.yaml")

	componentComponent := core.Component{
		Name:      "cloud-native",
		Source:    "https://github.com/timfpark/fabriakte-cloud-native",
		Method:    "git",
		Generator: "component",
	}

	err = Add(componentComponent)
	assert.Nil(t, err)

	helmComponent := core.Component{
		Name:      "elasticsearch",
		Source:    "https://github.com/helm/charts",
		Method:    "git",
		Path:      "stable/elasticsearch",
		Generator: "helm",
	}

	err = Add(helmComponent)
	assert.Nil(t, err)
}
