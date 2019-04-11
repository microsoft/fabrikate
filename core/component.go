package core

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"reflect"
	"strings"

	"github.com/kyokomi/emoji"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	yaml "github.com/timfpark/yaml"
)

type Component struct {
	Name          string              `yaml:"name" json:"name"`
	Config        ComponentConfig     `yaml:"-" json:"-"`
	Generator     string              `yaml:"generator,omitempty" json:"generator,omitempty"`
	Hooks         map[string][]string `yaml:"hooks,omitempty" json:"hooks,omitempty"`
	Serialization string              `yaml:"-" json:"-"`
	Source        string              `yaml:"source,omitempty" json:"source,omitempty"`
	Method        string              `yaml:"method,omitempty" json:"method,omitempty"`
	Path          string              `yaml:"path,omitempty" json:"path,omitempty"`
	Version       string              `yaml:"version,omitempty" json:"version,omitempty"`
	Branch        string              `yaml:"branch,omitempty" json:"branch,omitempty"`

	Repositories  map[string]string `yaml:"repositories,omitempty" json:"repositories,omitempty"`
	Subcomponents []Component       `yaml:"subcomponents,omitempty" json:"subcomponents,omitempty"`

	PhysicalPath string `yaml:"-" json:"-"`
	LogicalPath  string `yaml:"-" json:"-"`

	Manifest string `yaml:"-" json:"-"`
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
	*yaml.DefaultMapType = reflect.TypeOf(map[string]interface{}{})
	err = c.UnmarshalComponent("yaml", yaml.Unmarshal, &mergedComponent)

	if err != nil {
		err = c.UnmarshalComponent("json", json.Unmarshal, &mergedComponent)
		if err != nil {
			errorMessage := fmt.Sprintf("Error loading component in path %s", c.PhysicalPath)
			return mergedComponent, errors.Errorf(errorMessage)
		} else {
			mergedComponent.Serialization = "json"
		}
	} else {
		mergedComponent.Serialization = "yaml"
	}

	mergedComponent.PhysicalPath = c.PhysicalPath
	mergedComponent.LogicalPath = c.LogicalPath
	err = mergedComponent.Config.Merge(c.Config)

	return mergedComponent, err
}

func (c *Component) LoadConfig(environments []string) (err error) {
	for _, environment := range environments {
		if err := c.Config.MergeConfigFile(c.PhysicalPath, environment); err != nil {
			return err
		}
	}

	return c.Config.MergeConfigFile(c.PhysicalPath, "common")
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

func (c *Component) ExecuteHook(hook string) (err error) {
	if c.Hooks[hook] == nil {
		return nil
	}

	log.Info(emoji.Sprintf(":fishing_pole_and_fish: executing hooks for: %s", hook))

	for _, command := range c.Hooks[hook] {
		log.Info(emoji.Sprintf(":fishing_pole_and_fish: executing command: %s", command))
		if len(command) != 0 {
			cmd := exec.Command("bash", "-c", command)
			cmd.Dir = c.PhysicalPath
			out, err := cmd.Output()

			if err != nil {
				log.Info(emoji.Sprintf(":no_entry_sign: %s\n", err.Error()))
				if ee, ok := err.(*exec.ExitError); ok {
					log.Info(emoji.Sprintf(":no_entry_sign: hook command failed with: %s\n", ee.Stderr))
				}
				return err
			}

			if len(out) > 0 {
				outstring := emoji.Sprintf(":mag_right: %s\n", out)
				log.Info(strings.TrimSpace(outstring))
			}
		}
	}

	return nil
}

func (c *Component) BeforeGenerate() (err error) {
	return c.ExecuteHook("before-generate")
}

func (c *Component) AfterGenerate() (err error) {
	return c.ExecuteHook("after-generate")
}

func (c *Component) BeforeInstall() (err error) {
	return c.ExecuteHook("before-install")
}

func (c *Component) AfterInstall() (err error) {
	return c.ExecuteHook("after-install")
}

func (c *Component) InstallComponent(componentPath string) (err error) {
	if c.Method == "git" {
		componentsPath := fmt.Sprintf("%s/components", componentPath)
		if err := exec.Command("mkdir", "-p", componentsPath).Run(); err != nil {
			return err
		}

		subcomponentPath := path.Join(componentPath, c.RelativePathTo())
		if err = exec.Command("rm", "-rf", subcomponentPath).Run(); err != nil {
			return err
		}

		log.Println(emoji.Sprintf(":helicopter: installing component %s with git from %s", c.Name, c.Source))
		if err = CloneRepo(c.Source, c.Version, subcomponentPath, c.Branch); err != nil {
			return err
		}
	}

	return nil
}

func (c *Component) Install(componentPath string, generator Generator) (err error) {
	if err := c.BeforeInstall(); err != nil {
		return err
	}

	for _, subcomponent := range c.Subcomponents {
		if err := subcomponent.InstallComponent(componentPath); err != nil {
			return err
		}
	}

	if generator != nil {
		if err := generator.Install(c); err != nil {
			return err
		}
	}

	return c.AfterInstall()
}

func (c *Component) Generate(generator Generator) (err error) {
	if err := c.BeforeGenerate(); err != nil {
		return err
	}

	if generator != nil {
		c.Manifest, err = generator.Generate(c)
	} else {
		c.Manifest = ""
		err = nil
	}

	if err != nil {
		return err
	}

	return c.AfterGenerate()
}

type ComponentIteration func(path string, component *Component) (err error)

// IterateComponentTree is a general function used for iterating a deployment tree for installing, generating, etc.
func IterateComponentTree(startingPath string, environments []string, componentIteration ComponentIteration) (completedComponents []Component, err error) {
	queue := make([]Component, 0)

	component := Component{
		PhysicalPath: startingPath,
		LogicalPath:  "./",
		Config:       NewComponentConfig(startingPath),
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
		if err := component.LoadConfig(environments); err != nil {
			return nil, err
		}

		// 3. Call the passed componentIteration function on this component (install, generate, etc.)
		if err = componentIteration(component.PhysicalPath, &component); err != nil {
			return nil, err
		}

		completedComponents = append(completedComponents, component)

		configYAML, err := yaml.Marshal(component.Config)
		if err != nil {
			return nil, err
		}

		log.Debugf("Iterating component %s with config:\n%s", component.Name, string(configYAML))
		for _, subcomponent := range component.Subcomponents {
			subcomponent.Config = component.Config.Subcomponents[subcomponent.Name]

			subcomponentConfigYAML, err := yaml.Marshal(subcomponent.Config)
			if err != nil {
				return nil, err
			}

			log.Debugf("Iterating subcomponent '%s' with config:\n%s", subcomponent.Name, string(subcomponentConfigYAML))
			if len(subcomponent.Generator) == 0 && len(subcomponent.Source) > 0 {
				// This subcomponent is not inlined, so add it to the queue for iteration.

				subcomponent.PhysicalPath = path.Join(component.PhysicalPath, subcomponent.RelativePathTo())
				subcomponent.LogicalPath = path.Join(component.LogicalPath, subcomponent.Name)

				log.Debugf("Adding subcomponent '%s' to queue with physical path '%s' and logical path '%s'\n", subcomponent.Name, subcomponent.PhysicalPath, subcomponent.LogicalPath)
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

func (c *Component) Write() (err error) {
	var marshaledComponent []byte

	_ = os.Mkdir(c.PhysicalPath, os.ModePerm)

	if c.Serialization == "json" {
		marshaledComponent, err = json.MarshalIndent(c, "", "  ")
	} else {
		marshaledComponent, err = yaml.Marshal(c)
	}

	if err != nil {
		return err
	}

	filename := fmt.Sprintf("component.%s", c.Serialization)
	path := path.Join(c.PhysicalPath, filename)

	log.Info(emoji.Sprintf(":floppy_disk: Writing %s", path))

	return ioutil.WriteFile(path, marshaledComponent, 0644)
}
