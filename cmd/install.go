package cmd

import (
	"errors"
	"os/exec"

	"github.com/Microsoft/fabrikate/core"
	"github.com/Microsoft/fabrikate/generators"
	"github.com/kyokomi/emoji"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// Install installs the component at the given path and all of its subcomponents.
func Install(path string) (err error) {

	log.Info(emoji.Sprintf(":point_right: Initializing Helm"))
	if err = exec.Command("helm", "init", "--client-only").Run(); err != nil {
		return err
	}

	_, err = core.IterateComponentTree(path, []string{}, func(path string, component *core.Component) (err error) {
		log.Info(emoji.Sprintf(":point_right: starting install for component: %s", component.Name))

		var generator core.Generator

		switch component.Generator {
		case "helm":
			generator = &generators.HelmGenerator{}
		}

		if err := component.Install(path, generator); err != nil {
			return err
		}

		log.Info(emoji.Sprintf(":point_left: finished install for component: %s", component.Name))

		return err
	})

	if err == nil {
		log.Info(emoji.Sprintf(":raised_hands: finished install"))
	}

	return err
}

// installCmd represents the install command
var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Installs all of the remote components specified in the current deployment tree locally",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		PrintVersion()

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
