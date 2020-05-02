package cmd

import (
	"errors"
	"os/exec"

	"github.com/kyokomi/emoji"
	"github.com/microsoft/fabrikate/core"
	"github.com/microsoft/fabrikate/generators"
	"github.com/microsoft/fabrikate/logger"
	"github.com/spf13/cobra"
)

// Install implements the 'install' command.  It installs the component at the given path and all of
// its subcomponents by iterating the component subtree.
func Install(path string) (err error) {
	// Make sure host system contains all utils needed by Fabrikate
	requiredSystemTools := []string{"git", "helm", "sh", "curl"}
	for _, tool := range requiredSystemTools {
		path, err := exec.LookPath(tool)
		if err != nil {
			return err
		}
		logger.Info(emoji.Sprintf(":mag: Using %s: %s", tool, path))
	}

	// By default, Helm 2 used to have the "stable" repo. This is assumed
	// to exist in Fabrikate as a non-http repo.
	// See timeline/deprecation schedule: https://github.com/helm/charts
	logger.Info(emoji.Sprintf(":point_right: Adding stable repository"))
	if output, err := exec.Command("helm", "repo", "add", "stable", "https://kubernetes-charts.storage.googleapis.com").CombinedOutput(); err != nil {
		logger.Error(emoji.Sprintf(":no_entry_sign: %s: %s", err, output))
		return err
	}

	rootInit := func(startingPath string, environments []string, c core.Component) (component core.Component, err error) {
		return c.InstallRoot(startingPath, environments)
	}

	results := core.WalkComponentTree(path, []string{}, func(path string, component *core.Component) (err error) {
		logger.Info(emoji.Sprintf(":point_right: Starting install for component: %s", component.Name))

		var generator core.Generator

		switch component.ComponentType {
		case "helm":
			generator = &generators.HelmGenerator{}
		case "static":
			generator = &generators.StaticGenerator{}
		}

		// Load access tokens and add them to the global token list. Do not overwrite if already present
		if accessTokens, err := component.GetAccessTokens(); err != nil {
			return err
		} else if len(accessTokens) > 0 {
			for repo, token := range accessTokens {
				if _, exists := core.GitAccessTokens.Get(repo); !exists {
					core.GitAccessTokens.Set(repo, token)
				}
			}
		}

		if err := component.Install(path, generator); err != nil {
			return err
		}

		logger.Info(emoji.Sprintf(":point_left: Finished install for component: %s", component.Name))

		return err
	}, rootInit)

	components, err := core.SynchronizeWalkResult(results)
	if err != nil {
		return err
	}

	for _, component := range components {
		logger.Info(emoji.Sprintf(":white_check_mark: Installed successfully: %s", component.Name))
	}
	logger.Info(emoji.Sprintf(":raised_hands: Finished install"))

	return err
}

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Installs all of the remote components specified in the current deployment tree locally",
	Long: `Installs all of the remote components specified in the current deployment tree locally, iterating the 
component subtree from the current directory to do so.  Required to be executed before generate (if needed), such
that Fabrikate has all of the dependencies locally to use to generate the resource manifests.`,
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
}
