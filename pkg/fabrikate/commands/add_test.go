package commands

import (
	"os"
	"testing"

	"github.com/microsoft/fabrikate/pkg/fabrikate/core"
	"github.com/stretchr/testify/assert"
)

func TestAdd(t *testing.T) {

	// This test changes the cwd. Must change back so any tests following don't break
	cwd, err := os.Getwd()
	assert.Nil(t, err)
	defer func() {
		_ = os.Chdir(cwd)
	}()

	err = os.Chdir("../testdata/add")
	assert.Nil(t, err)

	_ = os.Remove("./component.yaml")

	componentComponent := core.Component{
		Name:          "cloud-native",
		Source:        "https://github.com/timfpark/fabrikate-cloud-native",
		Method:        "git",
		ComponentType: "component",
		Version:       "8ad79e73e0665e347e1553ad7ca32b6e590e007a",
	}

	err = Add(componentComponent)
	assert.Nil(t, err)

	helmComponent := core.Component{
		Name:          "elasticsearch",
		Source:        "https://github.com/helm/charts",
		Method:        "git",
		Path:          "stable/elasticsearch",
		ComponentType: "helm",
	}

	err = Add(helmComponent)
	assert.Nil(t, err)

	// Ensure the correct values are being added to the added subcomponents
	currentComponent := core.Component{}
	currentComponent, err = currentComponent.LoadComponent()
	assert.Nil(t, err)
	assert.EqualValues(t, componentComponent, currentComponent.Subcomponents[0])
	assert.EqualValues(t, helmComponent, currentComponent.Subcomponents[1])

	////////////////////////////////////////////////////////////////////////////////
	// Test adding a subcomponent
	////////////////////////////////////////////////////////////////////////////////
	subcomponentName := "My subcomponent"
	initialSource := "the initial URL; should not see this"
	newSource := "this should be the final value"
	err = componentComponent.AddSubcomponent(core.Component{
		Name:          subcomponentName,
		Source:        initialSource,
		Method:        "git",
		ComponentType: "component",
	})

	assert.Nil(t, err)
	assert.True(t, componentComponent.Subcomponents[0].Name == subcomponentName)
	assert.True(t, componentComponent.Subcomponents[0].Source == initialSource)

	err = componentComponent.AddSubcomponent(core.Component{
		Name:          subcomponentName,
		Source:        newSource,
		Method:        "git",
		ComponentType: "component",
	})

	// should still only have 1 subcomponent
	assert.Nil(t, err)
	assert.True(t, len(componentComponent.Subcomponents) == 1)
	assert.True(t, componentComponent.Subcomponents[0].Source == newSource)

	err = componentComponent.AddSubcomponent(core.Component{
		Name:          "this is a new name, so it should add a new subcomponent entry",
		Source:        newSource,
		Method:        "git",
		ComponentType: "component",
	})

	// there should be 2 subcomponents now
	assert.Nil(t, err)
	assert.True(t, len(componentComponent.Subcomponents) == 2)

	// Testing: ensure subcomponents sorted by name
	componentComponent.Subcomponents = []core.Component{}
	assert.True(t, len(componentComponent.Subcomponents) == 0)
	subcomponentA := core.Component{
		Name: "a",
	}
	subcomponentB := core.Component{
		Name: "b",
	}
	subcomponentC := core.Component{
		Name: "c",
	}

	// Add subcomponents in random order
	assert.Nil(t, componentComponent.AddSubcomponent(subcomponentC, subcomponentA, subcomponentB))

	// Subcomponent should be sorted by name
	assert.EqualValues(t, componentComponent.Subcomponents[0].Name, "a")
	assert.EqualValues(t, componentComponent.Subcomponents[1].Name, "b")
	assert.EqualValues(t, componentComponent.Subcomponents[2].Name, "c")
	////////////////////////////////////////////////////////////////////////////////
	//End adding a subcomponent
	////////////////////////////////////////////////////////////////////////////////
}
