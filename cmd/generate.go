package cmd

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"

	"github.com/Microsoft/fabrikate/core"
	"github.com/Microsoft/fabrikate/generators"
	"github.com/kyokomi/emoji"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func WriteGeneratedManifests(generationPath string, components []core.Component) (err error) {
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

		err = ioutil.WriteFile(componentYAMLFilePath, []byte(component.Manifest), 0644)
		if err != nil {
			return err
		}
	}

	return nil
}

func ValidateGeneratedManifests(generationPath string) (err error) {
	log.Println(emoji.Sprintf(":microscope: validating generated manifests in path %s", generationPath))
	output, err := exec.Command("kubectl", "apply", "--validate=true", "--dry-run", "--recursive", "-f", generationPath).Output()

	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			log.Errorf("helm template failed with: %s: output: %s\n", ee.Stderr, output)
			return err
		}
	}

	return nil
}

func Generate(startPath string, environment string) (components []core.Component, err error) {
	// Iterate through component tree and generate
	components, err = core.IterateComponentTree(startPath, environment, func(path string, component *core.Component) (err error) {

		var generator core.Generator
		switch component.Generator {
		case "helm":
			generator = &generators.HelmGenerator{}
		case "static":
			generator = &generators.StaticGenerator{}
		}

		return component.Generate(generator)
	})

	generationPath := path.Join(startPath, "generated", environment)

	if err = WriteGeneratedManifests(generationPath, components); err != nil {
		return nil, err
	}

	if err = ValidateGeneratedManifests(generationPath); err != nil {
		return nil, err
	}

	log.Info(emoji.Sprintf(":raised_hands: finished generate"))

	return components, err
}

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate <environment> [path]",
	Short: "Generates Kubernetes resource definitions from deployment definition.",
	Long:  `Generate produces Kubernetes resource definitions from deployment definition and an environment config.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 || len(args) > 2 {
			return errors.New("generate takes at least one or two arguments: 1) the name of the environment to be generated and 2) the path of the root of the defintion directory (defaults to the current directory).")
		}

		environment := args[0]

		path := "./"
		if len(args) > 1 {
			path = args[1]
		}

		_, err := Generate(path, environment)

		return err
	},
}

func init() {
	rootCmd.AddCommand(generateCmd)
}
