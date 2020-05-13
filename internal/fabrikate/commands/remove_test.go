package commands

import (
	"testing"

	"github.com/microsoft/fabrikate/internal/fabrikate/core"
	"github.com/stretchr/testify/assert"
)

func TestRemove(t *testing.T) {
	root := core.Component{
		Name: "root",
	}
	subcomponentC := core.Component{
		Name: "subcomponentC",
	}
	subcomponentA := core.Component{
		Name: "subcomponentA",
	}
	subcomponentB := core.Component{
		Name: "subcomponentB",
	}

	assert.Nil(t, root.AddSubcomponent(subcomponentC, subcomponentA, subcomponentB))
	assert.True(t, len(root.Subcomponents) == 3) // There should be 3 subcomponents

	assert.Nil(t, root.RemoveSubcomponent(subcomponentB))
	assert.True(t, len(root.Subcomponents) == 2)                  // There should be 2 subcomponents
	assert.True(t, root.Subcomponents[0].Name == "subcomponentA") // "subcomponentA" should be first after sorting
	assert.True(t, root.Subcomponents[1].Name == "subcomponentC") // "subcomponentC" should be second after sorting
}
