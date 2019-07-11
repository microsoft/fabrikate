package cmd

import (
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "fab",
	Short: "Scalable GitOps for Kubernetes clusters",
	Long:  "Scalable GitOps for Kubernetes clusters",

	PersistentPreRunE: func(cmd *cobra.Command, args []string) (err error) {
		verbose := cmd.Flag("verbose").Value.String()

		if verbose == "true" {
			log.SetLevel(log.DebugLevel)
		} else {
			log.SetLevel(log.InfoLevel)
		}

		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Error(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Use verbose output logs")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		if err != nil {
			log.Errorf("Getting home directory failed with: %s\n", err)
			os.Exit(1)
		}

		// Search config in home directory with name ".fab" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".fab")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		log.Debugf("Using config file: %s\n", viper.ConfigFileUsed())
	}
}
