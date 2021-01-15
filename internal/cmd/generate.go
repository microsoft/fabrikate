package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/kyokomi/emoji"
	"github.com/microsoft/fabrikate/internal/core"
	"github.com/microsoft/fabrikate/internal/generators"
	"github.com/microsoft/fabrikate/internal/logger"
	"github.com/spf13/cobra"
)

func writeGeneratedManifests(generationPath string, components []core.Component) (err error) {
	// Delete the old version, so we don't end up with a mishmash of two builds.
	os.RemoveAll(generationPath)

	for _, component := range components {
		componentGenerationPath := path.Join(generationPath, component.LogicalPath)
		if err = os.MkdirAll(componentGenerationPath, 0777); err != nil {
			return err
		}

		componentYAMLFilename := fmt.Sprintf("%s.yaml", component.Name)
		componentYAMLFilePath := path.Join(componentGenerationPath, componentYAMLFilename)

		logger.Info(emoji.Sprintf(":floppy_disk: Writing %s", componentYAMLFilePath))

		err = ioutil.WriteFile(componentYAMLFilePath, []byte(component.Manifest), 0644)
		if err != nil {
			return err
		}
	}

	return nil
}

func validateGeneratedManifests(generationPath string) (err error) {
	logger.Info(emoji.Sprintf(":microscope: Validating generated manifests in path %s", generationPath))
	if output, err := exec.Command("kubectl", "apply", "--validate=true", "--dry-run", "--recursive", "-f", generationPath).Output(); err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			logger.Error(fmt.Sprintf("Validating generated manifests failed with: %s: output: %s", ee.Stderr, output))
			return err
		}
	}

	return nil
}

// Generate implements the 'generate' command. It takes a set of environments and a validation flag
// and iterates through the component tree, generating components as it reaches them, and writing all
// of the generated manifests at the very end.
func Generate(startPath string, environments []string, validate bool) (components []core.Component, err error) {
	// Iterate through component tree and generate

	rootInit := func(startPath string, environments []string, c core.Component) (component core.Component, err error) {
		return c.UpdateComponentPath(startPath, environments)
	}

	results := core.WalkComponentTree(startPath, environments, func(path string, component *core.Component) (err error) {

		var generator core.Generator
		switch component.ComponentType {
		case "helm":
			generator = &generators.HelmGenerator{}
		case "static":
			generator = &generators.StaticGenerator{}
		case "":
			fallthrough
		case "component":
			// noop
		default:
			return fmt.Errorf(`invalid component type %v in component %+v`, component.ComponentType, component)
		}

		return component.Generate(generator)
	}, rootInit)

	components, err = core.SynchronizeWalkResult(results)

	if err != nil {
		return nil, err
	}

	environmentName := strings.Join(environments, "-")
	if len(environmentName) == 0 {
		environmentName = "common"
	}

	generationPath := path.Join(startPath, "generated", environmentName)

	if err = writeGeneratedManifests(generationPath, components); err != nil {
		return nil, err
	}

	if validate {
		if err = validateGeneratedManifests(generationPath); err != nil {
			return nil, err
		}
	}

	if err == nil {
		logger.Info(emoji.Sprintf(":raised_hands: Finished generate"))
	}

	return components, err
}

var generateCmd = &cobra.Command{
	Use:   "generate <config1> <config2> ... <configN>",
	Short: "Generates Kubernetes resource definitions from deployment definition.",
	Long: `Generate produces Kubernetes resource manifests from a deployment definition.

Where the generate command takes a list of the configurations that should be used to generate the resource
definitions for the deployment.  These configurations should be specified in priority order.  For example,
if you specified "prod azure east", prod's config would be applied first, and azure's config
would only be applied if they did not conflict with prod. Likewise, east's config would only be applied
if it did not conflict with prod or azure.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		PrintVersion()

		validation := cmd.Flag("validate").Value.String()
		_, err := Generate("./", args, validation == "true")

		return err
	},
}

func init() {
	generateCmd.PersistentFlags().Bool("validate", false, "Validate generated resource manifest YAML")
	rootCmd.AddCommand(generateCmd)
}
