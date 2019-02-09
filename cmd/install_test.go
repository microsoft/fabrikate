package cmd

import (
	"fmt"
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
