package core

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"

	"github.com/kyokomi/emoji"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"
)

type Component struct {
	Name   string
	Source string
	Method string

	Generator     string
	Subcomponents []Component
	Repo          string
	Path          string
	PhysicalPath  string
	LogicalPath   string
	Config        ComponentConfig

	Manifest string
}

type UnmarshalFunction func(in []byte, v interface{}) error

func UnmarshalFile(path string, unmarshalFunc UnmarshalFunction, obj interface{}) (err error) {
	_, err = os.Stat(path)
	if err != nil {
		return err
	}

	marshaled, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	log.Info(emoji.Sprintf(":floppy_disk: Loading %s", path))

	return unmarshalFunc(marshaled, obj)
}

func (c *Component) UnmarshalComponent(marshaledType string, unmarshalFunc UnmarshalFunction, component *Component) error {
	componentFilename := fmt.Sprintf("component.%s", marshaledType)
	componentPath := path.Join(c.PhysicalPath, componentFilename)

	return UnmarshalFile(componentPath, unmarshalFunc, component)
}

func (c *Component) LoadComponent() (mergedComponent Component, err error) {
	err = c.UnmarshalComponent("yaml", yaml.Unmarshal, &mergedComponent)
	if err != nil {
		err = c.UnmarshalComponent("json", json.Unmarshal, &mergedComponent)
		if err != nil {
			errorMessage := fmt.Sprintf("Error loading %s: %s", c.PhysicalPath, err)
			log.Errorln(errorMessage)
			return mergedComponent, errors.Errorf(errorMessage)
		}
	}

	mergedComponent.PhysicalPath = c.PhysicalPath
	mergedComponent.LogicalPath = c.LogicalPath
	mergedComponent.Config.Merge(c.Config)

	return mergedComponent, nil
}

func (c *Component) UnmarshalConfig(environment string, marshaledType string, unmarshalFunc UnmarshalFunction, config *ComponentConfig) error {
	configFilename := fmt.Sprintf("config/%s.%s", environment, marshaledType)
	configPath := path.Join(c.PhysicalPath, configFilename)

	return UnmarshalFile(configPath, unmarshalFunc, config)
}

func (c *Component) MergeConfigFile(environment string) (err error) {
	var componentConfig ComponentConfig

	err = c.UnmarshalConfig(environment, "yaml", yaml.Unmarshal, &componentConfig)
	if err != nil {
		err = c.UnmarshalConfig(environment, "json", json.Unmarshal, &componentConfig)
		if err != nil {
			return nil
		}
	}

	c.Config.Merge(componentConfig)

	return nil
}

func (c *Component) LoadConfig(environment string) (err error) {
	if err := c.MergeConfigFile(environment); err != nil {
		return err
	}

	return c.MergeConfigFile("common")
}

func (c *Component) RelativePathTo() string {
	if c.Method == "git" {
		return fmt.Sprintf("components/%s", c.Name)
	} else if c.Source != "" {
		return c.Name
	} else {
		return "./"
	}
}

func (c *Component) Install(componentPath string) (err error) {
	for _, subcomponent := range c.Subcomponents {
		if subcomponent.Method == "git" {
			componentsPath := fmt.Sprintf("%s/components", componentPath)
			if err := exec.Command("mkdir", "-p", componentsPath).Run(); err != nil {
				return err
			}

			subcomponentPath := path.Join(componentPath, subcomponent.RelativePathTo())
			if err = exec.Command("rm", "-rf", subcomponentPath).Run(); err != nil {
				return err
			}

			log.Println(emoji.Sprintf(":helicopter: installing component %s with git from %s", subcomponent.Name, subcomponent.Source))
			if err = exec.Command("git", "clone", subcomponent.Source, subcomponentPath).Run(); err != nil {
				return err
			}
		}
	}

	return nil
}

type ComponentIteration func(path string, component *Component) (err error)

// IterateComponentTree is a general function used for iterating a deployment tree for installing, generating, etc.
func IterateComponentTree(startingPath string, environment string, componentIteration ComponentIteration) (completedComponents []Component, err error) {
	queue := make([]Component, 0)

	component := Component{
		PhysicalPath: startingPath,
		LogicalPath:  "./",
		Config: ComponentConfig{
			Config:        make(map[string]interface{}),
			Subcomponents: make(map[string]ComponentConfig),
		},
	}

	queue = append(queue, component)
	completedComponents = make([]Component, 0)

	// Iterate the deployment tree using a queued breadth first algorithm:
	for len(queue) != 0 {
		component := queue[0]
		queue = queue[1:]

		// 1. Parse the component at that path into a Component
		component, err := component.LoadComponent()
		if err != nil {
			return nil, err
		}

		// 2. Load the config for this Component
		if err := component.LoadConfig(environment); err != nil {
			return nil, err
		}

		// 3. Call the passed componentIteration function on this component (install, generate, etc.)
		if err = componentIteration(component.PhysicalPath, &component); err != nil {
			return nil, err
		}

		completedComponents = append(completedComponents, component)

		for _, subcomponent := range component.Subcomponents {
			subcomponent.Config = component.Config.Subcomponents[subcomponent.Name]
			if len(subcomponent.Source) > 0 {
				// This subcomponent is not inlined, so add it to the queue for iteration.

				subcomponent.PhysicalPath = path.Join(component.PhysicalPath, subcomponent.RelativePathTo())
				subcomponent.LogicalPath = path.Join(component.LogicalPath, subcomponent.Name)

				log.Debugf("adding subcomponent '%s' to queue with physical path '%s' and logical path '%s'\n", subcomponent.Name, subcomponent.PhysicalPath, subcomponent.LogicalPath)
				queue = append(queue, subcomponent)
			} else {
				// This subcomponent is inlined, so call the componentIteration function on this component.

				subcomponent.PhysicalPath = component.PhysicalPath
				subcomponent.LogicalPath = component.LogicalPath

				if err = componentIteration(subcomponent.PhysicalPath, &subcomponent); err != nil {
					return nil, err
				}

				completedComponents = append(completedComponents, subcomponent)
			}
		}
	}

	return completedComponents, nil
}
