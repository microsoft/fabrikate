package cmd

import (
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/microsoft/fabrikate/core"
	"github.com/stretchr/testify/assert"
	"github.com/timfpark/yaml"
)

const GENERATION_PATH = "../testdata/kustomize/generated"

var TEST_COMPONENT core.Component = core.Component{
	Name:        "test-component",
	LogicalPath: "infra",
}

func TestSetDefaultEmptyKustomization(t *testing.T) {
	k := kustomization{}
	k.setDefaultEmptyKustomization()

	assert.Equal(t, k.Kind, defaultKind)
	assert.Equal(t, k.APIVersion, defaultAPIVersion)
	assert.Equal(t, k.Resources, []string{})
}

func TestAddKustomizationResource(t *testing.T) {
	k := kustomization{}

	k.addKustomizationResource(TEST_COMPONENT)

	componentYAMLFilename := fmt.Sprintf("%s.yaml", TEST_COMPONENT.Name)
	filePath := path.Join(TEST_COMPONENT.LogicalPath, componentYAMLFilename)

	assert.Equal(t, 1, len(k.Resources))
	assert.Equal(t, filePath, k.Resources[0])
}

func TestWriteKustomizationFile(t *testing.T) {
	k := kustomization{}
	k.setDefaultEmptyKustomization()
	k.addKustomizationResource(TEST_COMPONENT)
	err := os.MkdirAll(GENERATION_PATH, 0777)
	assert.Nil(t, err)

	kustomizationBytes, err := yaml.Marshal(k)
	assert.Nil(t, err)

	err = writeKustomizationFile(GENERATION_PATH, kustomizationBytes)
	assert.Nil(t, err)
}

func TestCreateKustomizationFile(t *testing.T) {
	components := []core.Component{TEST_COMPONENT}
	err := os.MkdirAll(GENERATION_PATH, 0777)
	assert.Nil(t, err)

	err = createKustomizationFile(GENERATION_PATH, components)

	assert.Nil(t, err)
}
