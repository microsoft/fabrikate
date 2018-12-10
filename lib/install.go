package lib

import (
	"github.com/Microsoft/marina/models"
)

func Install(startingPath string) (results []models.ComponentResult, err error) {
	return IterateComponentTree(startingPath, func(path string, component *models.Component) (result string, err error) {
		err = component.Install(path)
		return "", err
	})
}
