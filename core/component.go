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
	"sort"
	"strings"
	"sync"

	"github.com/kyokomi/emoji"
	"github.com/microsoft/fabrikate/logger"
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

	logger.Info(emoji.Sprintf(":floppy_disk: Loading %s", path))

	return unmarshalFunc(marshaled, output)
}

// UnmarshalComponent finds and unmarshal the component.<format> of a component using the
// provided `unmarshalFunc` function.
func (c *Component) UnmarshalComponent(serializationType string, unmarshalFunc unmarshalFunction, component *Component) error {
	logger.Debug(fmt.Sprintf("Attempting to unmarshal %s for component '%s'", serializationType, c.Name))

	componentFilename := fmt.Sprintf("component.%s", serializationType)
	componentPath := path.Join(c.PhysicalPath, componentFilename)

	err := UnmarshalFile(componentPath, unmarshalFunc, component)
	component.Serialization = serializationType

	return err
}

func (c *Component) applyDefaultsAndMigrations() error {
	if len(c.Generator) > 0 {
		logger.Warn(emoji.Sprintf(":boom: DEPRECATION WARNING: Field 'generator' has been deprecated and will be removed in version v1.0.0; Update component '%s' to use 'type' in place of 'generator'", c.Name))
		c.ComponentType = c.Generator
	}

	if len(c.Repositories) > 0 {
		logger.Warn(emoji.Sprintf(":boom: DEPRECATION WARNING: Field `repositories` has been deprecrated and will be removed in version v1.0.0; Update component '%s' to use `method: helm`, `source: <helm_repo_url>`, and `path: <chart_name>` and remove `repositories`", c.Name))
	}

	if len(c.ComponentType) == 0 {
		c.ComponentType = "component"
	}

	return nil
}

// LoadComponent loads a component definition in either YAML or JSON formats.
func (c *Component) LoadComponent() (loadedComponent Component, err error) {

	// If success or loading or parsing the yaml component failed for reasons other than it didn't exist, return.
	if err = c.UnmarshalComponent("yaml", yaml.Unmarshal, &loadedComponent); err != nil && !os.IsNotExist(err) {
		return loadedComponent, err
	}

	// If YAML component definition did not exist, try JSON.
	if err != nil {
		if err = c.UnmarshalComponent("json", json.Unmarshal, &loadedComponent); err != nil {
			if !os.IsNotExist(err) {
				return loadedComponent, err
			}

			errorMessage := fmt.Sprintf("Error loading component in path %s", c.PhysicalPath)
			return loadedComponent, errors.New(errorMessage)
		}
	}

	if err = loadedComponent.applyDefaultsAndMigrations(); err != nil {
		return loadedComponent, err
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
		// The component is in filesystem
		return c.Source
	}

	return "./"
}

// ExecuteHook executes the passed hook
func (c *Component) ExecuteHook(hook string) (err error) {
	if c.Hooks[hook] == nil {
		return nil
	}

	for _, command := range c.Hooks[hook] {
		logger.Info(emoji.Sprintf(":fishing_pole_and_fish: Executing command in hook '%s' for component '%s': %s", hook, c.Name, command))
		if len(command) != 0 {
			cmd := exec.Command("sh", "-c", command)
			cmd.Dir = c.PhysicalPath
			output, err := cmd.CombinedOutput()
			if err != nil {
				logger.Error(emoji.Sprintf(":no_entry_sign: Error occurred in hook '%s' for component '%s'\n%s: %s", hook, c.Name, err, output))
				return err
			}
			if len(output) > 0 {
				outstring := emoji.Sprintf(":mag_right: Completed hook '%s' for component '%s':\n%s", hook, c.Name, output)
				logger.Trace(strings.TrimSpace(outstring))
			}
		}
	}

	return nil
}

// beforeGenerate executes the 'before-generate' hook (if any) of the component.
func (c *Component) beforeGenerate() (err error) {
	return c.ExecuteHook("before-generate")
}

// afterGenerate executes the 'after-generate' hook (if any) of the component.
func (c *Component) afterGenerate() (err error) {
	return c.ExecuteHook("after-generate")
}

// beforeInstall executes the 'before-install' hook (if any) of the component.
func (c *Component) beforeInstall() (err error) {
	return c.ExecuteHook("before-install")
}

// afterInstall executes the 'after-install' hook (if any) of the component.
func (c *Component) afterInstall() (err error) {
	return c.ExecuteHook("after-install")
}

// InstallComponent installs the component (if needed) utilizing its Method.
// This is only used to install 'components', Generators handle the installation
// of 'non-components' (eg; helm/static). Therefore the only installation needed
// for any component is when ComponentType == "component"|""  and Method ==
// "git"
func (c *Component) InstallComponent(componentPath string) (err error) {
	if c.ComponentType == "component" {
		if c.Method == "git" {
			// ensure `components` dir exists
			componentsPath := path.Join(componentPath, "components")
			if err := os.MkdirAll(componentsPath, 0777); err != nil {
				return err
			}

			// delete the subcomponent if previously installed
			subcomponentPath := path.Join(componentPath, c.RelativePathTo())
			if err = os.RemoveAll(subcomponentPath); err != nil {
				return err
			}

			logger.Info(emoji.Sprintf(":helicopter: Installing component '%s' with git from '%s'", c.Name, c.Source))
			if err = Git.CloneRepo(c.Source, c.Version, subcomponentPath, c.Branch); err != nil {
				return err
			}
			return nil
		}
	}

	return nil
}

// InstallSingleComponent installs the given component
func (c *Component) InstallSingleComponent(componentPath string, generator Generator) (err error) {
	if err := c.beforeInstall(); err != nil {
		return err
	}

	if err := c.applyDefaultsAndMigrations(); err != nil {
		return err
	}

	if err := c.InstallComponent(componentPath); err != nil {
		return err
	}

	if generator != nil {
		if err := generator.Install(c); err != nil {
			return err
		}
	}

	return c.afterInstall()
}

// Install encapsulates the install lifecycle of a component including before-install,
// installation, and after-install hooks.
func (c *Component) Install(componentPath string, generator Generator) (err error) {
	if err := c.beforeInstall(); err != nil {
		return err
	}

	// Install subcomponents
	for _, subcomponent := range c.Subcomponents {
		if err = subcomponent.applyDefaultsAndMigrations(); err != nil {
			return err
		}
		if err := subcomponent.InstallComponent(componentPath); err != nil {
			return err
		}
	}

	// Install self
	if generator != nil {
		if err := generator.Install(c); err != nil {
			return err
		}
	}

	return c.afterInstall()
}

// Generate encapsulates the generate lifecycle of a component including before-generate,
// generation, and after-generate hooks.
func (c *Component) Generate(generator Generator) (err error) {
	if err := c.beforeGenerate(); err != nil {
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

	return c.afterGenerate()
}

type componentIteration func(path string, component *Component) (err error)

type rootComponentInit func(startingPath string, environments []string, c Component) (component Component, err error)

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
func WalkComponentTree(startingPath string, environments []string, iterator componentIteration, rootInit rootComponentInit) <-chan WalkResult {
	queue := make(chan Component)    // components enqueued to be 'visited' (ie; walked over)
	results := make(chan WalkResult) // To pass WalkResults to
	walking := sync.WaitGroup{}      // Keep track of all nodes being worked on

	// Prepares `component` by loading/de-serializing the component.yaml/json and configs
	// Note: this is only needed for non-inlined components
	prepareComponent := func(c Component) Component {
		logger.Debug(fmt.Sprintf("Preparing component '%s'", c.Name))
		// 1. Parse the component at that path into a Component
		c, err := c.LoadComponent()
		if err != nil {
			results <- WalkResult{Error: err}
		}

		// 2. Load the config for this Component
		if err = c.LoadConfig(environments); err != nil {
			results <- WalkResult{Error: err}
		}
		return c
	}

	// Enqueue the given component
	enqueue := func(c Component) {
		// Increment working counter; MUST happen BEFORE sending to queue or race condition can occur
		walking.Add(1)
		logger.Debug(fmt.Sprintf("Adding subcomponent '%s' to queue with physical path '%s' and logical path '%s'\n", c.Name, c.PhysicalPath, c.LogicalPath))
		queue <- c
	}

	// Mark a component as visited and report it back as a result; decrements the walking counter
	markAsVisited := func(c *Component) {
		results <- WalkResult{Component: c}
		walking.Done()
	}

	// Main worker thread to enqueue root node, wait, and close the channel once all nodes visited
	go func() {
		// Manually enqueue the first root component

		rootComponent := prepareComponent(Component{
			PhysicalPath: startingPath,
			LogicalPath:  "./",
			Config:       NewComponentConfig(startingPath),
		})

		// Init rootComponent
		rootComponent, err := rootInit(startingPath, environments, rootComponent)

		if err != nil {
			results <- WalkResult{Error: err}
		} else {
			enqueue(rootComponent)
		}

		// Close results channel once all nodes visited
		walking.Wait()
		close(results)
	}()

	// Worker thread to pull from queue and call the iterator
	go func() {
		for queuedComponent := range queue {
			go func(c Component) {
				// Decrement working counter; Must happen AFTER the subcomponents are enqueued
				defer markAsVisited(&c)

				// Call the iterator
				err := iterator(c.PhysicalPath, &c)
				if err != nil {
					results <- WalkResult{Error: err}
				}

				// Range over subcomponents; preparing and enqueuing
				for _, subcomponent := range c.Subcomponents {
					// Prep component config
					subcomponent.Config = c.Config.Subcomponents[subcomponent.Name]

					if err = subcomponent.applyDefaultsAndMigrations(); err != nil {
						results <- WalkResult{Error: err}
					}

					// Do not add to the queue if component or subcomponent is Disabled.
					if subcomponent.Config.Disabled {
						logger.Info(emoji.Sprintf(":prohibited: Subcomponent '%s' is disabled", subcomponent.Name))
						continue
					}

					// Depending if the subcomponent is inlined or not; prepare the component to either load
					// config/path info from filesystem (non-inlined) or inherit from parent (inlined)
					if subcomponent.ComponentType == "component" || subcomponent.ComponentType == "" {
						// This subcomponent is not inlined, so set the paths to their relative positions and prepare the configs
						subcomponent.PhysicalPath = path.Join(subcomponent.RelativePathTo(), subcomponent.Path)
						if !filepath.IsAbs(subcomponent.RelativePathTo()) {
							subcomponent.PhysicalPath = path.Join(c.PhysicalPath, subcomponent.PhysicalPath)
						}
						subcomponent.LogicalPath = path.Join(c.LogicalPath, subcomponent.Name)
						subcomponent = prepareComponent(subcomponent)
					} else {
						// This subcomponent is inlined, so it inherits paths from parent and no need to prepareComponent().
						subcomponent.PhysicalPath = c.PhysicalPath
						subcomponent.LogicalPath = c.LogicalPath
					}

					logger.Debug(fmt.Sprintf("Adding subcomponent '%s' to queue with physical path '%s' and logical path '%s'\n", subcomponent.Name, subcomponent.PhysicalPath, subcomponent.LogicalPath))
					enqueue(subcomponent)
				}
			}(queuedComponent)
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
	componentPath := path.Join(c.PhysicalPath, filename)

	logger.Info(emoji.Sprintf(":floppy_disk: Writing '%s'", componentPath))

	return ioutil.WriteFile(componentPath, marshaledComponent, 0644)
}

// AddSubcomponent adds the provided subcomponents to a component.
// If the subcomponents Name matches an existing entry, the existing entry is overwritten.
// If the subcomponents Name does not match, a new subcomponent entry is created.
func (c *Component) AddSubcomponent(subcomponents ...Component) (err error) {
	// Index all existing components based on name and add the new component
	// Warning: this will remove any duplicates with the same name if present
	nameComponentMap := map[string]Component{}
	for _, subcomponent := range c.Subcomponents {
		nameComponentMap[subcomponent.Name] = subcomponent
	}
	for _, subcomponent := range subcomponents {
		nameComponentMap[subcomponent.Name] = subcomponent
	}

	// Re-add all subcomponents so no named collisions occur
	c.Subcomponents = []Component{}
	for _, subcomponent := range nameComponentMap {
		c.Subcomponents = append(c.Subcomponents, subcomponent)
	}

	// Sort by subcomponent name to ensure order is maintained
	c.sortSubcomponents()

	return nil
}

// RemoveSubcomponent takes in a variadic amount of subcomponents and attempts to remove them from the component.
// If a subcomponent is found with the same name, it is removed. If not, it is a noop.
func (c *Component) RemoveSubcomponent(subcomponents ...Component) (err error) {
	// Index all existing components based on name and then delete necessary ones based on name
	// Warning: this will remove any duplicates with the same name if present
	nameComponentMap := map[string]Component{}
	for _, subcomponent := range c.Subcomponents {
		nameComponentMap[subcomponent.Name] = subcomponent
	}

	// Delete all components matching .Name of subcomponents
	for _, subcomponent := range subcomponents {
		delete(nameComponentMap, subcomponent.Name)
	}

	// Re-add all subcomponents so no named collisions occur
	c.Subcomponents = []Component{}
	for _, subcomponent := range nameComponentMap {
		c.Subcomponents = append(c.Subcomponents, subcomponent)
	}

	// Sort by subcomponent name to ensure order is maintained
	c.sortSubcomponents()

	return nil
}

// sortSubcomponents sorts a components subcomponents in decending alphabetical order.
func (c *Component) sortSubcomponents() {
	// Sort by subcomponent name to ensure order is maintained
	sort.Slice(c.Subcomponents, func(i, j int) bool {
		return c.Subcomponents[i].Name < c.Subcomponents[j].Name
	})
}

// GetAccessTokens attempts to find an access.yaml file in the same physical directory of the component.
// Un-marshalling if found,
func (c *Component) GetAccessTokens() (tokens map[string]string, err error) {
	// If access.yaml is found in same directory of component.yaml, see if c.Source is in the map and use the value as accessToken
	accessYamlPath := path.Join(c.PhysicalPath, "access.yaml")
	if err = UnmarshalFile(accessYamlPath, yaml.Unmarshal, &tokens); os.IsNotExist(err) {
		// If the file is not found, return an empty map with no error
		return map[string]string{}, nil
	} else if err != nil {
		logger.Error(emoji.Sprintf(":no_entry_sign: Error unmarshalling access.yaml in '%s'", accessYamlPath))
		return nil, err
	}

	// Attempt to load env variables listed in access.yaml
	for repo, envVar := range tokens {
		token := os.Getenv(envVar)
		if token == "" {
			// Give warning that failed to load env var; but continue and attempt clone
			msg := fmt.Sprintf("Component '%s' attempted to load environment variable '%s'; but is either not set or an empty string. Components with source '%s' may fail to install", c.Name, envVar, repo)
			logger.Warn(emoji.Sprintf(":no_entry_sign: %s", msg))
		} else {
			tokens[repo] = token
		}
	}
	return tokens, err
}

// InstallRoot installs the root component
func (c Component) InstallRoot(startingPath string, environments []string) (root Component, err error) {
	logger.Debug(fmt.Sprintf("Install root component'%s'", c.Name))

	if c.Method != "git" {
		return c, err
	}

	// Install the root
	if err := c.InstallSingleComponent(startingPath, nil); err != nil {
		return c, err
	}

	return c.UpdateComponentPath(startingPath, environments)
}

// UpdateComponentPath updates the component path if it required installing another component
func (c Component) UpdateComponentPath(startingPath string, environments []string) (root Component, err error) {
	logger.Debug(fmt.Sprintf("Update component path'%s'", c.Name))

	if c.Method != "git" {
		return c, err
	}

	if c.ComponentType == "component" || c.ComponentType == "" {
		relativePath := c.RelativePathTo()
		c.PhysicalPath = path.Join(relativePath, c.Path)
		if !filepath.IsAbs(c.RelativePathTo()) {
			c.PhysicalPath = path.Join(startingPath, c.PhysicalPath)
		}

		c, err = c.LoadComponent()
		if err != nil {
			return c, err
		}

		if err = c.LoadConfig(environments); err != nil {
			return c, err
		}
	}
	return c, err
}
