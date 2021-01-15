package core

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/microsoft/fabrikate/internal/installable"
	"github.com/timfpark/yaml"
)

func Load(componentDirectory string) (c Component, err error) {
	// calculate possible component.<yaml|json> filepaths
	componentBase := "component"
	componentAbsDir, err := filepath.Abs(componentDirectory)
	if err != nil {
		return c, fmt.Errorf(`failed to compute absolute path of component "%v": %w`, componentDirectory, err)
	}
	yamlPath := path.Join(componentAbsDir, componentBase+".yaml")
	jsonPath := path.Join(componentAbsDir, componentBase+".json")
	_, yamlErr := os.Stat(yamlPath)
	_, jsonErr := os.Stat(jsonPath)
	isYAML := yamlErr == nil
	isJSON := jsonErr == nil

	// find the correct path
	var componentPath string
	if isYAML && isJSON {
		return c, fmt.Errorf(`only one of component.<yaml|json> can exist per component, both found in "%v"`, componentAbsDir)
	} else if isYAML {
		componentPath = yamlPath
	} else if isJSON {
		componentPath = jsonPath
	} else {
		return c, fmt.Errorf(`component.<yaml|json> not found in "%v"`, componentAbsDir)
	}

	// read the file and unmarshal into found extension type
	componentBytes, err := ioutil.ReadFile(componentPath)
	if err != nil {
		return c, fmt.Errorf(`failed to read component file "%v": %w`, componentPath, err)
	}
	if isYAML {
		if err := yaml.Unmarshal(componentBytes, &c); err != nil {
			return c, fmt.Errorf(`failed to unmarshal component.yaml at "%v": %w`, componentPath, err)
		}
	} else if isJSON {
		if err := json.Unmarshal(componentBytes, &c); err != nil {
			return c, fmt.Errorf(`failed to unmarshal component.json at "%v": %w`, componentPath, err)
		}
	} else {
		return c, fmt.Errorf("invalid component extension")
	}

	return c, err
}

func (c Component) ToInstallable() (installer installable.Installable, err error) {
	switch c.Method {
	case "git":
		installer = installable.Git{
			URL:    c.Source,
			Branch: c.Branch,
			SHA:    c.Version,
		}
	case "helm":
		installer = installable.Helm{
			URL:     c.Source,
			Chart:   c.Path,
			Version: c.Version,
		}
	case "local":
		installer = installable.Local{
			Root: c.Path,
		}
	case "":
		// noop
	default:
		return installer, fmt.Errorf(`unsupported method "%v" in component "%+v"`, c.Method, c)
	}

	return installer, err
}

func echo(level int, message interface{}) {
	decorator := "-"
	switch level {
	case 0:
		decorator = ">"
	case 1:
		decorator = "\u2192" // right arrow
	case 2:
		decorator = "+"
	}
	// indent := strings.Repeat("    ", level)
	indent := strings.Repeat("\t", level)
	fmt.Printf("%v%v %v\n", indent, decorator, message)
}

func Install(startPath string) error {
	echo(0, fmt.Sprintf(`Staring Fabrikate installation at: "%v"`, startPath))
	c, err := Load(startPath)
	if err != nil {
		fmt.Printf("Error loading component in path \"%v\": %v\n", startPath, err)
	}
	visited, err := install([]Component{c}, []Component{})
	if err != nil {
		return err
	}

	echo(0, "Installation report:")
	for _, c := range visited {
		echo(1, fmt.Sprintf("%v", c.Name))
	}

	return nil
}

func install(queue []Component, visited []Component) ([]Component, error) {
	//----------------------------------------------------------------------------
	// base case
	if len(queue) == 0 {
		return visited, nil
	}

	//----------------------------------------------------------------------------
	// recursive case
	first, rest := queue[0], queue[1:]

	echo(0, fmt.Sprintf(`Installing component: "%v"`, first.Name))
	echo(1, "Adding subcomponents to queue")
	for _, sub := range first.Subcomponents {
		rest = append(rest, sub)
		echo(2, fmt.Sprintf(`Added component to queue: "%v"`, sub.Name))
	}

	echo(1, "Executing hook: Before-Install")
	if err := first.beforeInstall(); err != nil {
		echo(2, fmt.Errorf(`error running "before-install" hook: %w`, err))
	}

	installer, err := first.ToInstallable()
	if err != nil {
		return visited, fmt.Errorf(`error installing component "%v": %w`, first.Name, err)
	}
	if installer != nil {
		echo(1, "Validating coordinate")
		if err := installer.Validate(); err != nil {
			return visited, fmt.Errorf(`validation failed for component coordinate "%v": %w`, installer, err)
		}

		echo(1, "Computing installation path")
		installPath, err := installer.GetInstallPath()
		if err != nil {
			return visited, err
		}
		echo(2, fmt.Sprintf(`Installation path: "%v"`, installPath))

		echo(1, "Cleaning previous installation")
		if _, exists := os.Stat(installPath); exists == nil {
			echo(2, fmt.Sprintf(`Previous installation found at %v`, installPath))
			if err := os.RemoveAll(installPath); err != nil {
				return visited, fmt.Errorf(`error removing existing installation "%v": %w`, installPath, err)
			}
		}

		echo(1, "Installing")
		if err := installer.Install(); err != nil {
			return visited, fmt.Errorf(`error installing component "%v": %w`, first.Name, err)
		}
		echo(2, fmt.Sprintf(`Installed component to: "%v"`, installPath))

		// add remote components to the queue
		componentType := strings.ToLower(first.ComponentType)
		if componentType == "" || componentType == "component" {
			remoteComponentPath := path.Join(installPath, first.Path)
			echo(2, fmt.Sprintf(`Adding fetched remote component to queue: "%v"`, remoteComponentPath))
			echo(3, fmt.Sprintf(`Loading component: "%v"`, remoteComponentPath))
			remoteComponent, err := Load(remoteComponentPath)
			if err != nil {
				return visited, fmt.Errorf(`error loading component from path "%v": %w`, installPath, err)
			}
			echo(4, fmt.Sprintf(`Loaded component: "%v"`, remoteComponent.Name))
			rest = append(rest, remoteComponent)
			echo(3, fmt.Sprintf(`Added remote component to queue: "%v"`, remoteComponent.Name))
		}
	}

	echo(1, "Executing hook: After-Install")
	if err := first.afterInstall(); err != nil {
		echo(2, fmt.Errorf(`error running "after-install" hook: %w`, err))
	}

	visited = append(visited, first)
	echo(1, "Installation complete")

	return install(rest, visited)
}

// func (c Component) toGeneratable() (generator generatable.Generatable, err error) {
// 	installer, err := c.toInstallable()
// 	if err != nil {
// 		return generator, err
// 	}
// 	installPath, err := installer.GetInstallPath()
// 	if err != nil {
// 		return generator, err
// 	}

// 	switch c.ComponentType {
// 	case "helm":
// 		generator = generatable.Helm{
// 			ChartPath: installPath,
// 		}
// 	case "static":
// 		generator = generatable.Static{
// 			ManifestPath: installPath,
// 		}
// 	case "":
// 		fallthrough // same as "component"
// 	case "component":
// 		// noop
// 	default:
// 		return generator, fmt.Errorf(`unsupported type "%v" in component "%+v"`, c.ComponentType, c)
// 	}

// 	return generator, err
// }

type iterator = func(c Component) error

func Iterate(startPath string, visit iterator) ([]Component, error) {
	// Load starting component
	component, err := Load(startPath)
	if err != nil {
		return nil, fmt.Errorf(`failed to load component at "%v": %w`, startPath, err)
	}
	// Initialize the tree tracking properties of the component
	component.PhysicalPath = startPath
	component.LogicalPath = fmt.Sprintf("%v", os.PathSeparator)

	return iterate(visit, []Component{component}, []Component{})
}

func iterate(visit iterator, queue []Component, visited []Component) ([]Component, error) {
	//----------------------------------------------------------------------------
	// Base case
	if len(queue) == 0 {
		return visited, nil
	}

	//----------------------------------------------------------------------------
	// Recursive case
	first, rest := queue[0], queue[1:]
	environments := []string{"common"}
	if err := first.LoadConfig(environments); err != nil {
		return visited, fmt.Errorf(`error loading configuration "%v" for component %+v: %w`, environments, first, err)
	}

	// Visit the component
	if err := visit(first); err != nil {
		return visited, fmt.Errorf(`error visiting component %+v during component iteration: %w`, first, err)
	}

	// Add all children to the queue
	for _, child := range first.Subcomponents {
		childType := strings.ToLower(child.ComponentType)
		// If of type `component`, load the remote location on disk and enqueue.
		// Else, it is inline, so just enqueue the inlined component
		if childType == "" || childType == "component" {
			installer, err := child.ToInstallable()
			if err != nil {
				return visited, fmt.Errorf(`error converting subcomponent %+v to installable: %w`, child, err)
			}
			physicalPath, err := installer.GetInstallPath()
			if err != nil {
				return visited, fmt.Errorf(`error computing installation path of subcomponent %+v: %w`, child, err)
			}
			remote, err := Load(physicalPath)
			if err != nil {
				return visited, fmt.Errorf(`error loading subcomponent in path "%+v": %w`, physicalPath, err)
			}
			// Remote components have their own component paths
			remote.PhysicalPath = physicalPath
			remote.LogicalPath = path.Join(first.LogicalPath, child.Name)
			rest = append(rest, remote)
		} else {
			// Inlined components inherit component paths from the parent
			child.PhysicalPath = first.PhysicalPath
			child.LogicalPath = path.Join(first.LogicalPath, first.Name)
			rest = append(rest, child)
		}
	}

	return iterate(visit, rest, visited)
}
