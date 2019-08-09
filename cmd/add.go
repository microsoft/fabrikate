package cmd

import (
	"errors"
	"os"
	"strings"

	"github.com/kyokomi/emoji"
	"github.com/microsoft/fabrikate/core"
	log "github.com/sirupsen/logrus"
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

	err = component.AddSubcomponent(subcomponent)
	if err != nil {
		return err
	}

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

		// If method not "git", set branch to zero value
		method := cmd.Flag("method").Value.String()
		branch := cmd.Flag("branch").Value.String()
		if cmd.Flags().Changed("method") && method != "git" {
			// Warn users if they explicitly set --branch that the config is being removed
			if cmd.Flags().Changed("branch") {
				log.Warn(emoji.Sprintf(":exclamation: Non 'git' --method and explicit --branch specified. Removing --branch configuration of 'branch: %s'", branch))
			}
			branch = ""
		}

		component := core.Component{
			Name:          args[0],
			Source:        cmd.Flag("source").Value.String(),
			Method:        method,
			Branch:        branch,
			Path:          cmd.Flag("path").Value.String(),
			ComponentType: cmd.Flag("type").Value.String(),
		}

		return Add(component)
	},
}

func init() {
	addCmd.PersistentFlags().String("source", "", "Source for this component")
	addCmd.PersistentFlags().String("method", "git", "Method to use to fetch this component")
	addCmd.PersistentFlags().String("branch", "master", "Branch of git repo to use; noop when method != 'git'")
	addCmd.PersistentFlags().String("path", "./", "Path of git repo to use")
	addCmd.PersistentFlags().String("type", "component", "Type of this component")

	rootCmd.AddCommand(addCmd)
}
