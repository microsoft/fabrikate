package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRelativePathToGitComponent(t *testing.T) {
	subcomponent := Subcomponent{
		Name:   "efk",
		Method: "git",
		Source: "https://github.com/Microsoft/marina-elasticsearch-fluentd-kibana",
	}

	assert.Equal(t, subcomponent.RelativePathTo(), "components/efk")
}

func TestRelativePathToDirectoryComponent(t *testing.T) {
	subcomponent := Subcomponent{
		Name:   "infra",
		Source: "./infra",
	}

	assert.Equal(t, subcomponent.RelativePathTo(), "infra")
}
