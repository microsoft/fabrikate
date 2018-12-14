package cmd

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/Microsoft/marina/core"
	"github.com/Microsoft/marina/generators"
	"github.com/spf13/cobra"
)

func Generate(startPath string, environment string) (components []core.Component, err error) {
	// Iterate through component tree and generate
	components, err = core.IterateComponentTree(startPath, environment, func(path string, component *core.Component) (err error) {
		if component.Type == "helm" {
			component.Definition, err = generators.GenerateHelmComponent(component)
		}
		return err
	})

	// Delete the old version, so we don't end up with a mishmash of two builds.
	generationPath := path.Join(startPath, "generated", environment)
	os.RemoveAll(generationPath)

	// TODO: need to push component yaml out to {path}/generated directory
	for _, component := range components {
		componentGenerationPath := path.Join(generationPath, component.LogicalPath)
		os.MkdirAll(componentGenerationPath, 0755)

		componentYAMLFilename := fmt.Sprintf("%s.yaml", component.Name)
		componentYAMLFilePath := path.Join(componentGenerationPath, componentYAMLFilename)

		ioutil.WriteFile(componentYAMLFilePath, []byte(component.Definition), 0644)
	}

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

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// generateCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// generateCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
