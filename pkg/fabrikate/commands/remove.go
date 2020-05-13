package commands

import (
	"errors"
	"os"
	"strings"

	"github.com/microsoft/fabrikate/pkg/fabrikate/core"
	"github.com/spf13/cobra"
)

// Remove implements the `remove` command. Taking in a list of subcomponent names, this function
// will load the root component and attempt to remove any subcomponents with names matching
// those provided.
func Remove(subcomponent core.Component) (err error) {
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

	err = component.RemoveSubcomponent(subcomponent)
	if err != nil {
		return err
	}

	return component.Write()
}

var removeCmd = &cobra.Command{
	Use:   "remove <component-name>",
	Short: "Removes a subcomponent from the current component.",
	Long: `Removes a subcomponent from the current component.

example:

$ fab remove fabrikate-cloud-native
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) <= 0 {
			return errors.New("'remove' takes one or more component-name arguments")
		}

		component := core.Component{
			Name: args[0],
		}

		return Remove(component)
	},
}

func init() {
	rootCmd.AddCommand(removeCmd)
}
