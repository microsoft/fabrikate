package cmd

import (
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/microsoft/fabrikate/util"
	"github.com/stretchr/testify/assert"
	"github.com/timfpark/yaml"
)

func TestInstallJSON(t *testing.T) {
	componentDir := "../testdata/install"
	cwd, err := os.Getwd()
	assert.Nil(t, err)
	defer func() {
		assert.Nil(t, os.Chdir(cwd))
		assert.Nil(t, util.UninstallComponents(componentDir))
	}()

	// Change cwd to component directory
	assert.Nil(t, os.Chdir(componentDir))
	assert.Nil(t, Install("./"))
}

func TestInstallYAML(t *testing.T) {
	componentDir := "../testdata/install-yaml"
	cwd, err := os.Getwd()
	assert.Nil(t, err)
	defer func() {
		assert.Nil(t, os.Chdir(cwd))
		assert.Nil(t, util.UninstallComponents(componentDir))
	}()

	// Change cwd to component directory
	assert.Nil(t, os.Chdir(componentDir))
	assert.Nil(t, Install("./"))
}

func TestInstallWithHooks(t *testing.T) {
	componentDir := "../testdata/install-hooks"
	cwd, err := os.Getwd()
	assert.Nil(t, err)
	defer func() {
		assert.Nil(t, os.Chdir(cwd))
		assert.Nil(t, util.UninstallComponents(componentDir))
	}()

	// Change cwd to component directory
	assert.Nil(t, os.Chdir(componentDir))

	assert.Nil(t, Install("./"))
}

func TestInstallPrivateComponent(t *testing.T) {
	componentDir := "../testdata/install-private"
	cwd, err := os.Getwd()
	assert.Nil(t, err)
	defer func() {
		assert.Nil(t, os.Chdir(cwd))
		assert.Nil(t, util.UninstallComponents(componentDir))
	}()

	// Change cwd to component directory
	assert.Nil(t, os.Chdir(componentDir))

	// Should fail with no environment var set to personal_access_token
	assert.NotNil(t, Install("./"))
	assert.Nil(t, os.Chdir("./"))

	// If a personal access token exists, assume its correct and Install should succeed
	if _, exists := os.LookupEnv("personal_access_token"); exists {
		assert.Nil(t, Install("./"))
	} else {
		assert.NotNil(t, Install("./"))
	}
}

func TestInstallHelmMethod(t *testing.T) {
	componentDir := "../testdata/install-helm"
	cwd, err := os.Getwd()
	assert.Nil(t, err)
	defer func() {
		assert.Nil(t, os.Chdir(cwd))
		assert.Nil(t, util.UninstallComponents(componentDir))
	}()

	// Change cwd to component directory
	assert.Nil(t, os.Chdir(componentDir))
	assert.Nil(t, Install("./"))

	// Grafana chart should be version 3.7.0
	grafanaChartYaml := path.Join("helm_repos", "grafana", "Chart.yaml")
	grafanaChartBytes, err := ioutil.ReadFile(grafanaChartYaml)
	assert.Nil(t, err)
	type helmChart struct {
		Version string
		Name    string
	}
	grafanaChart := helmChart{}
	assert.Nil(t, yaml.Unmarshal(grafanaChartBytes, &grafanaChart))
	assert.EqualValues(t, "grafana", grafanaChart.Name)
	assert.EqualValues(t, "3.7.0", grafanaChart.Version)
}
