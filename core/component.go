package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"reflect"
	"strings"
	"sync"

	"github.com/kyokomi/emoji"
	log "github.com/sirupsen/logrus"
	"github.com/timfpark/yaml"
)

// Component documentation: https://github.com/microsoft/fabrikate/blob/master/docs/component.md
type Component struct {
	Name          string              `yaml:"name" json:"name"`
	Config        ComponentConfig     `yaml:"-" json:"-"`
	ComponentType string              `yaml:"type,omitempty" json:"type,omitempty"`
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

type unmarshalFunction func(in []byte, v interface{}) error

// UnmarshalFile is an unmarshal wrapper which reads in any file from `path` and attempts to
// unmarshal to `output` using the `unmarshalFunc`.
func UnmarshalFile(path string, unmarshalFunc unmarshalFunction, output interface{}) (err error) {
	_, err = os.Stat(path)
	if err != nil {
		return err
	}

	marshaled, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	log.Info(emoji.Sprintf(":floppy_disk: Loading %s", path))

	return unmarshalFunc(marshaled, output)
}

// UnmarshalComponent finds and unmarshal the component.<format> of a component using the
// provided `unmarshalFunc` function.
func (c *Component) UnmarshalComponent(marshaledType string, unmarshalFunc unmarshalFunction, component *Component) error {
	componentFilename := fmt.Sprintf("component.%s", marshaledType)
	componentPath := path.Join(c.PhysicalPath, componentFilename)

	return UnmarshalFile(componentPath, unmarshalFunc, component)
}

// LoadComponent loads the component at c.PhysicalPath
func (c *Component) LoadComponent() (mergedComponent Component, err error) {
	*yaml.DefaultMapType = reflect.TypeOf(map[string]interface{}{})
	err = c.UnmarshalComponent("yaml", yaml.Unmarshal, &loadedComponent)

	if err != nil {
		err = c.UnmarshalComponent("json", json.Unmarshal, &loadedComponent)
		if err != nil {
			errorMessage := fmt.Sprintf("Error loading component in path %s", c.PhysicalPath)
			return loadedComponent, errors.New(errorMessage)
		}
		loadedComponent.Serialization = "json"
	} else {
		loadedComponent.Serialization = "yaml"
	}

	if len(loadedComponent.Generator) > 0 {
		log.Println(emoji.Sprintf(":boom: DEPRECATION WARNING: Field 'generator' has been deprecated and will be removed in version 0.7.0."))
		log.Println(emoji.Sprintf(":boom: DEPRECATION WARNING: Update your component definition to use 'type' in place of 'generator'."))

		loadedComponent.ComponentType = loadedComponent.Generator
	}

	loadedComponent.PhysicalPath = c.PhysicalPath
	loadedComponent.LogicalPath = c.LogicalPath
	err = loadedComponent.Config.Merge(c.Config)

	return loadedComponent, err
}

// LoadConfig loads and merges the config specified by the passed set of environments.
func (c *Component) LoadConfig(environments []string) (err error) {
	for _, environment := range environments {
		if err := c.Config.MergeConfigFile(c.PhysicalPath, environment); err != nil {
			return err
		}
	}

	return c.Config.MergeConfigFile(c.PhysicalPath, "common")
}

// RelativePathTo returns the relative filesystem path where this component should be.
// If the method the component is retrieved is `git`: the convention "components/<component.Name>" is used
// If the method not git but the component has a `source`, that value is used
func (c *Component) RelativePathTo() string {
	if c.Method == "git" {
		return fmt.Sprintf("components/%s", c.Name)
	} else if c.Source != "" {
		return c.Source
	}

	return "./"
}

// ExecuteHook executes the passed hook
func (c *Component) ExecuteHook(hook string) (err error) {
	if c.Hooks[hook] == nil {
		return nil
	}

	log.Info(emoji.Sprintf(":fishing_pole_and_fish: executing hooks for: %s", hook))

	for _, command := range c.Hooks[hook] {
		log.Info(emoji.Sprintf(":fishing_pole_and_fish: executing command: %s", command))
		if len(command) != 0 {
			cmd := exec.Command("sh", "-c", command)
			cmd.Dir = c.PhysicalPath
			out, err := cmd.Output()

			if err != nil {
				cwd, _ := os.Getwd()
				log.Error(fmt.Sprintf("ERROR IN: %s", cwd))
				log.Error(emoji.Sprintf(":no_entry_sign: %s\n", err.Error()))
				if ee, ok := err.(*exec.ExitError); ok {
					log.Error(emoji.Sprintf(":no_entry_sign: hook command failed with: %s\n", ee.Stderr))
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

// BeforeGenerate executes the 'before-generate' hook (if any) of the component.
func (c *Component) BeforeGenerate() (err error) {
	return c.ExecuteHook("before-generate")
}

// AfterGenerate executes the 'after-generate' hook (if any) of the component.
func (c *Component) AfterGenerate() (err error) {
	return c.ExecuteHook("after-generate")
}

// BeforeInstall executes the 'before-install' hook (if any) of the component.
func (c *Component) BeforeInstall() (err error) {
	return c.ExecuteHook("before-install")
}

// AfterInstall executes the 'after-install' hook (if any) of the component.
func (c *Component) AfterInstall() (err error) {
	return c.ExecuteHook("after-install")
}

// InstallComponent installs the component (if needed) utilizing its Method.
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

// Install encapsulates the install lifecycle of a component including before-install,
// installation, and after-install hooks.
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

// Generate encapsulates the generate lifecycle of a component including before-generate,
// generation, and after-generate hooks.
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

type componentIteration func(path string, component *Component) (err error)

// WalkResult is what WalkComponentTree returns.
// Will contain either a Component OR an Error (Error is nillable; meaning both fields can be nil)
type WalkResult struct {
	Component *Component
	Error     error
}

// WalkComponentTree asynchronously walks a component tree starting at `startingPath` and calls
// `iterator` on every node in the tree in Breadth First Order.
//
// Returns a channel of WalkResult which can either have a Component or an Error (Error is nillable)
//
// Same level ordering is not ensured; any nodes on the same tree level can be visited in any order.
// Parent->Child ordering is ensured; A parent is always visited via `iterator` before the children are visited.
func WalkComponentTree(startingPath string, environments []string, iterator componentIteration) <-chan WalkResult {
	queue := make(chan Component)    // components enqueued to be 'visited' (ie; walked over)
	results := make(chan WalkResult) // To pass WalkResults to
	walking := sync.WaitGroup{}      // Keep track of all nodes being worked on

	// Prepares `component` by loading/de-serializing the component.yaml/json and configs
	// Note: this is only needed for non-inlined components
	prepareComponent := func(component Component) Component {
		// 1. Parse the component at that path into a Component
		component, err := component.LoadComponent()
		results <- WalkResult{Error: err}

		// 2. Load the config for this Component
		results <- WalkResult{Error: component.LoadConfig(environments)}
		return component
	}

	// Enqueue the given component
	enqueue := func(component Component) {
		// Increment working counter; MUST happen BEFORE sending to queue or race condition can occur
		walking.Add(1)
		log.Debugf("adding subcomponent '%s' to queue with physical path '%s' and logical path '%s'\n", component.Name, component.PhysicalPath, component.LogicalPath)
		queue <- component
	}

	// Mark a component as visited and report it back as a result; decrements the walking counter
	markAsVisited := func(component *Component) {
		results <- WalkResult{Component: component}
		walking.Done()
	}

	// Main worker thread to enqueue root node, wait, and close the channel once all nodes visited
	go func() {
		// Manually enqueue the first root component
		enqueue(prepareComponent(Component{
			PhysicalPath: startingPath,
			LogicalPath:  "./",
			Config:       NewComponentConfig(startingPath),
		}))

		// Close results channel once all nodes visited
		walking.Wait()
		close(results)
	}()

	// Worker thread to pull from queue and call the iterator
	go func() {
		for component := range queue {
			go func(component Component) {
				// Decrement working counter; Must happen AFTER the subcomponents are enqueued
				defer markAsVisited(&component)

				// Call the iterator
				results <- WalkResult{Error: iterator(component.PhysicalPath, &component)}

				// Range over subcomponents; preparing and enqueuing
				for _, subcomponent := range component.Subcomponents {
					// Prep component config
					subcomponent.Config = component.Config.Subcomponents[subcomponent.Name]

					// Depending if the subcomponent is inlined or not; prepare the component to either load
					// config/path info from filesystem (non-inlined) or inherit from parent (inlined)
					isNotInlined := (len(subcomponent.ComponentType) == 0 || subcomponent.ComponentType == "component") && len(subcomponent.Source) > 0
					if isNotInlined {
						// This subcomponent is not inlined, so set the paths to their relative positions and prepare the configs
						if subcomponent.Path == "" {
							// Standard component
							subcomponent.PhysicalPath = path.Join(component.PhysicalPath, subcomponent.RelativePathTo())
						} else {
							// Components Source points to a Fabrikate component library; concat Path to get to target component
							physicalPath := path.Join(subcomponent.RelativePathTo(), subcomponent.Path)
							if !filepath.IsAbs(subcomponent.RelativePathTo()) {
								physicalPath = path.Join(component.PhysicalPath, physicalPath)
							}
							subcomponent.PhysicalPath = physicalPath
						}
						subcomponent.LogicalPath = path.Join(component.LogicalPath, subcomponent.Name)
						subcomponent = prepareComponent(subcomponent)
					} else {
						// This subcomponent is inlined, so it inherits paths from parent and no need to prepareComponent().
						subcomponent.PhysicalPath = component.PhysicalPath
						subcomponent.LogicalPath = component.LogicalPath
					}

					log.Debugf("adding subcomponent '%s' to queue with physical path '%s' and logical path '%s'\n", subcomponent.Name, subcomponent.PhysicalPath, subcomponent.LogicalPath)
					enqueue(subcomponent)
				}
			}(component)
		}
	}()

	return results
}

// SynchronizeWalkResult will synchronize a channel of WalkResult to a list of visited Components.
// It will return on the first Error encountered; returning the visited Components up until then and the error
func SynchronizeWalkResult(results <-chan WalkResult) (components []Component, err error) {
	components = []Component{}
	for result := range results {
		if result.Error != nil {
			return components, result.Error
		} else if result.Component != nil {
			components = append(components, *result.Component)
		}
	}
	return components, err
}

// Write serializes a component to YAML (default) or JSON (chosen via c.Serialization) at c.PhysicalPath
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
