package generators

import (
	"github.com/microsoft/fabrikate/core"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStaticGenerator_Generate(t *testing.T) {
	component := core.Component{
		Name:         "foo",
		Path:         "",
		PhysicalPath: "../testdata/invaliddir",
	}

	generator := &StaticGenerator{}
	_, err := generator.Generate(&component)
	assert.NotNil(t, err)
}
