package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/Microsoft/fabrikate/core"
	"github.com/Microsoft/fabrikate/generators"
	"github.com/kyokomi/emoji"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func writeGeneratedManifests(generationPath string, components []core.Component) (err error) {
	// Delete the old version, so we don't end up with a mishmash of two builds.
	os.RemoveAll(generationPath)

	for _, component := range components {
		componentGenerationPath := path.Join(generationPath, component.LogicalPath)
		err := os.MkdirAll(componentGenerationPath, 0755)
		if err != nil {
			return err
		}

		componentYAMLFilename := fmt.Sprintf("%s.yaml", component.Name)
		componentYAMLFilePath := path.Join(componentGenerationPath, componentYAMLFilename)

		log.Info(emoji.Sprintf(":floppy_disk: Writing %s", componentYAMLFilePath))

		err = ioutil.WriteFile(componentYAMLFilePath, []byte(component.Manifest), 0644)
		if err != nil {
			return err
		}
	}

	return nil
}

func validateGeneratedManifests(generationPath string) (err error) {
	log.Println(emoji.Sprintf(":microscope: validating generated manifests in path %s", generationPath))
	output, err := exec.Command("kubectl", "apply", "--validate=true", "--dry-run", "--recursive", "-f", generationPath).Output()

	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			log.Errorf("validating generated manifests failed with: %s: output: %s\n", ee.Stderr, output)
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
	results := core.WalkComponentTree(startPath, environments, func(path string, component *core.Component) (err error) {

		var generator core.Generator
		switch component.Generator {
		case "helm":
			generator = &generators.HelmGenerator{}
		case "static":
			generator = &generators.StaticGenerator{}
		}

		return component.Generate(generator)
	})

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
		log.Info(emoji.Sprintf(":raised_hands: finished generate"))
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
