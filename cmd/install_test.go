package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInstall(t *testing.T) {
	err := Install("../test/fixtures/install")

	assert.Nil(t, err)
}

func TestInstallYaml(t *testing.T) {
	err := Install("../test/fixtures/install-yaml")

	assert.Nil(t, err)
}
