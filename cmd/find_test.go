package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/google/go-github/v28/github"
)

func TestGetFabrikateComponents(t *testing.T) {
	githubCodeResults := []github.CodeResult{}
	paths := []string{
		"definitions/fabrikate-prometheus-grafana/README.md",
		"samples/kafka-strimzi-portworx/config/README.md",
		"definitions/linkerd/README.md",
		"definitions/linkerd/component.yaml",
		"samples/kafka-strimzi-portworx/config/common.yaml",
	}

	for _, path := range paths{
		var p string = path
		githubCodeResults = append(githubCodeResults, github.CodeResult{Path: &p})
	}

	components := GetFabrikateComponents(githubCodeResults)
	assert.Equal(t, 2, len(components))
}

func TestGetFabrikateComponentsEmpty(t *testing.T) {
	githubCodeResults := []github.CodeResult{}

	components := GetFabrikateComponents(githubCodeResults)
	assert.Equal(t, 0, len(components))
}