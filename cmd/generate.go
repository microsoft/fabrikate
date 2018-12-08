package cmd

import (
	"errors"

	"github.com/spf13/cobra"
)

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate <config>",
	Short: "Generates Kubernetes resource definitions from deployment definition.",
	Long:  `Generate produces Kubernetes resource definitions from deployment definition and an environment config.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return errors.New("Generate requires an environment to be passed to the generate command")
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
