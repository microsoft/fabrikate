package cmd

import (
	"errors"
	"os"
	"strings"

	"github.com/microsoft/fabrikate/core"
	"github.com/spf13/cobra"
)

// Add implements the 'add' command in Fabrikate.  It takes a spec for the new subcomponent, loads
// the previous component (if any), adds the subcomponent, and serializes the new component back out.
func Add(subcomponent core.Component) (err error) {
	component := core.Component{
		PhysicalPath: "./",
		LogicalPath:  "",
	}

	component, err = component.LoadComponent()
	if err != nil {
		path, err := os.Getwd()
		if err != nil {
			return err
		}

		pathParts := strings.Split(path, "/")

		component = core.Component{
			Name:          pathParts[len(pathParts)-1],
			Serialization: "yaml",
		}
	}

	component.Subcomponents = append(component.Subcomponents, subcomponent)

	return component.Write()
}

var addCmd = &cobra.Command{
	Use:   "add <component-name> --source <component-source> [--type component] [--method git] [--path .]",
	Short: "Adds a subcomponent to the current component (or the component specified by the passed path).",
	Long: `Adds a subcomponent to the current component (or the component specified by the passed path).

source: where the component lives (either a local path or remote http(s) endpoint)
type: the type of component (component (default), helm, or static)
method: method used to fetch the component (git (default))
path: the path to the component that this subcomponent should be added to.

example:

$ fab add cloud-native --source https://github.com/microsoft/fabrikate-definitions --path definitions/fabrikate-cloud-native
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("'add' takes one or more key=value arguments")
		}

		component := core.Component{
			Name:          args[0],
			Source:        cmd.Flag("source").Value.String(),
			Method:        cmd.Flag("method").Value.String(),
			Branch:        cmd.Flag("branch").Value.String(),
			Path:          cmd.Flag("path").Value.String(),
			ComponentType: cmd.Flag("type").Value.String(),
		}

		return Add(component)
	},
}

func init() {
	addCmd.PersistentFlags().String("source", "", "Source for this component")
	addCmd.PersistentFlags().String("method", "git", "Method to use to fetch this component (default: git)")
	addCmd.PersistentFlags().String("branch", "master", "Branch of git repo to use (default: master)")
	addCmd.PersistentFlags().String("path", "", "Path of git repo to use (default: ./)")
	addCmd.PersistentFlags().String("type", "component", "Type of this component (default: component)")

	rootCmd.AddCommand(addCmd)
}
