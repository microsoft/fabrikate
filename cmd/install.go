package cmd

import (
	"errors"
	"fmt"

	"github.com/Microsoft/marina/core"
	"github.com/Microsoft/marina/generators"
	"github.com/spf13/cobra"
)

func Install(path string) (err error) {
	_, err = core.IterateComponentTree(path, "", func(path string, component *core.Component) (err error) {
		fmt.Printf("--> starting install for component: %s\n", component.Name)
		if err := component.Install(path); err != nil {
			return err
		}

		if component.Type == "helm" {
			err = generators.InstallHelmComponent(component)
		}

		if err == nil {
			fmt.Printf("<-- finished install for component: %s\n", component.Name)
		}

		return err
	})

	return err
}

// installCmd represents the install command
var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Installs all of the remote components specified in the current deployment tree locally",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		path := "./"

		if len(args) == 1 {
			path = args[0]
		}

		if len(args) > 1 {
			return errors.New("install takes zero or one arguments: the path to the root of the definition tree (defaults to current directory)")
		}

		return Install(path)
	},
}

func init() {
	rootCmd.AddCommand(installCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// installCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// installCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
