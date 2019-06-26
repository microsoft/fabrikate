package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInstallJSON(t *testing.T) {
	err := Install("../test/fixtures/install")

	if ee, ok := err.(*exec.ExitError); ok {
		fmt.Printf("TestInstallJSON failed with error %s\n", ee.Stderr)
	}

	assert.Nil(t, err)
}

func TestInstallYAML(t *testing.T) {
	err := Install("../test/fixtures/install-yaml")

	if ee, ok := err.(*exec.ExitError); ok {
		fmt.Printf("TestInstallYAML failed with error %s\n", ee.Stderr)
	}

	assert.Nil(t, err)
}

func TestInstallWithHooks(t *testing.T) {
	err := Install("../test/fixtures/install-hooks")

	assert.Nil(t, err)
}

func TestInstallPrivateComponent(t *testing.T) {
	cwd, err := os.Getwd()
	assert.Nil(t, err)
	defer func() {
		_ = os.Chdir(cwd)
	}()

	// Change cwd to component directory
	assert.Nil(t, os.Chdir("../test/fixtures/install-private"))

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
