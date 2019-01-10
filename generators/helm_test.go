package generators

import (
	"testing"

	"github.com/Microsoft/fabrikate/core"

	"github.com/stretchr/testify/assert"
)

func TestMakeHelmRepoPath(t *testing.T) {
	component := &core.Component{
		Name:      "istio",
		Generator: "helm",
		Path:      "./chart/istio-1.0.5",
	}

	assert.Equal(t, MakeHelmRepoPath(component), "chart/istio-1.0.5")
}
