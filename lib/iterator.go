package lib

import (
	"io/ioutil"
	"os"
	"path"

	"github.com/Microsoft/marina/models"
	"github.com/pkg/errors"
)

/*
IterateComponentTree is a general service used for iterating a deployment tree for installing, generating, etc.

It takes a starting path that is expected to have a component.json in it. It is assumed to be an error in this step of any path that
is pushed onto the queue when component.json does not exist in the path.

For each component path in the queue, it parses the component at that path into a Component, calls componentFunc on that, and then
for each subcomponent specified it determines if it is a simple subdirectory of if it (<subcomponent path>) is an installed component
in components and requires a two level path addition (components/<subcomponent name>).

Note: Because it is going a breadth first search, this enables an install operation to install components before the iteration discovers
they are missing.

*/

type ComponentFunc func(path string, component *models.Component) (result string, err error)

func IterateComponentTree(startingPath string, componentFunc ComponentFunc) (results []models.ComponentResult, err error) {
	queue := make([]models.ComponentResult, 0)
	initialComponentResult := models.ComponentResult{PhysicalPath: startingPath, LogicalPath: "./"}
	queue = append(queue, initialComponentResult)
	for len(queue) != 0 {
		currentComponentResult := queue[0]
		queue = queue[1:]

		componentJSONPath := path.Join(currentComponentResult.PhysicalPath, "component.json")

		if _, err := os.Stat(componentJSONPath); os.IsNotExist(err) {
			return nil, errors.Errorf("Component expected at path %s not found", componentJSONPath)
		}

		componentJSON, err := ioutil.ReadFile(componentJSONPath)
		if err != nil {
			return nil, err
		}

		component, err := models.ParseComponentFromJson(componentJSON)
		if err != nil {
			return nil, err
		}

		currentComponentResult.Result, err = componentFunc(currentComponentResult.PhysicalPath, component)
		if err != nil {
			return nil, err
		}

		results = append(results, currentComponentResult)

		for _, subcomponent := range component.Subcomponents {
			subcomponentResult := models.ComponentResult{
				PhysicalPath: path.Join(currentComponentResult.PhysicalPath, subcomponent.RelativePathTo()),
				LogicalPath:  path.Join(currentComponentResult.LogicalPath, subcomponent.Name),
			}

			queue = append(queue, subcomponentResult)
		}
	}

	return results, nil
}
