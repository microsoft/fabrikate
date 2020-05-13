package commands

import (
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/google/uuid"
	"github.com/microsoft/fabrikate/internal/fabrikate/util"
	"github.com/microsoft/fabrikate/pkg/encoding/yaml"
	"github.com/stretchr/testify/assert"
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

	// Installing again should not cause errors
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

	// Installing again should not cause errors
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

	// Installing again should not cause errors
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

// Test to cover https://github.com/microsoft/fabrikate/issues/261
// Tests the calling of Install when the helm client isn't initialized and
// attempts to get updates from `http://127.0.0.1:8879/charts` which is fails
// in most cases as people do not typically run helm servers locally.
func TestInstallWithoutHelmInitialized(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	assert.Nil(t, err)
	helmDir := path.Join(homeDir, ".helm")

	if _, err := os.Stat(helmDir); !os.IsNotExist(err) {
		// Move helm dir to a temporary location to simulate uninitialized helm client
		randomTmpName, err := uuid.NewRandom()
		assert.Nil(t, err)
		tmpDir := path.Join(homeDir, randomTmpName.String())
		assert.Nil(t, os.Rename(helmDir, tmpDir))

		// Ensure it is moved back to normal
		defer func() {
			assert.Nil(t, os.RemoveAll(helmDir))
			assert.Nil(t, os.Rename(tmpDir, helmDir))
		}()
	}

	componentDir := "../testdata/install-helm-fix-261-dep-update-bug"
	cwd, err := os.Getwd()
	assert.Nil(t, err)
	defer func() {
		assert.Nil(t, os.Chdir(cwd))
		assert.Nil(t, util.UninstallComponents(componentDir))
	}()

	// Change cwd to component directory and install
	assert.Nil(t, os.Chdir(componentDir))
	assert.Nil(t, Install("./"))
}

func TestGenerateHelmRepoAlias(t *testing.T) {
	componentDir := "../testdata/repo-alias"
	cwd, err := os.Getwd()
	assert.Nil(t, err)
	defer func() {
		assert.Nil(t, os.Chdir(cwd))
		assert.Nil(t, util.UninstallComponents(componentDir))
	}()

	// Change cwd to component directory
	assert.Nil(t, os.Chdir(componentDir))
	assert.Nil(t, Install("./"))

	assert.Nil(t, err)
}
