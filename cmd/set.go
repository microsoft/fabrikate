package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/Microsoft/fabrikate/core"
	"github.com/spf13/cobra"
)

func SplitPathValuePairs(pathValuePairStrings []string) (pathValuePairs []core.PathValuePair, err error) {
	for _, pathValuePairString := range pathValuePairStrings {
		pathValuePairParts := strings.Split(pathValuePairString, "=")

		if len(pathValuePairParts) != 2 {
			return pathValuePairs, fmt.Errorf("%s is not a properly formated configuration key/value pair.", pathValuePairString)
		}

		pathValuePair := core.PathValuePair{
			Path:  strings.Split(pathValuePairParts[0], "."),
			Value: pathValuePairParts[1],
		}

		pathValuePairs = append(pathValuePairs, pathValuePair)
	}

	return pathValuePairs, nil
}

func Set(environment string, subcomponent string, pathValuePairStrings []string, newconfigfail bool) (err error) {
	subcomponentPath := []string{}
	if len(subcomponent) > 0 {
		subcomponentPath = strings.Split(subcomponent, ".")
	}

	componentConfig := core.NewComponentConfig(".")

	pathValuePairs, err := SplitPathValuePairs(pathValuePairStrings)

	if err != nil {
		return err
	}

	if err := componentConfig.Load(environment); err != nil {
		return err
	}

	newConfigError := "New configuration was specified and the --newconfigfail switch is on."		
	
	for _, pathValue := range pathValuePairs {
			if newconfigfail {
				if (!componentConfig.HasSubcomponentConfig(subcomponentPath)) {
					return errors.New(newConfigError)
				} else {
					sc := componentConfig.GetSubcomponentConfig(subcomponentPath)

					if(!sc.HasComponentConfig(pathValue.Path)) {
						return errors.New(newConfigError)
					}
				}
			}

		componentConfig.SetConfig(subcomponentPath, pathValue.Path, pathValue.Value)
	}

	return componentConfig.Write(environment)
}

var subcomponent string
var environment string
var newconfigfail bool

var setCmd = &cobra.Command{
	Use:   "set <config> [--subcomponent subcomponent] <path1>=<value1> <path2>=<value2> ...",
	Short: "Sets a config value for a component for a particular config environment in the Fabrikate definition.",
	Long: `Sets a config value for a component for a particular config environment in the Fabrikate definition.
eg.
$ fab set --environment prod data.replicas=4 username="ops"

Sets the value of 'data.replicas' equal to 4 and 'username' equal to 'ops' in the 'prod' config for the current component.

$ fab set --subcomponent "myapp" endpoint="east-db" 

Sets the value of 'endpoint' equal to 'east-db' in the 'common' config (the default) for subcomponent 'myapp'.

$ fab set --subcomponent "myapp.mysubcomponent" data.replicas=5 

Sets the subkey "replicas" in the key 'data' equal to 5 in the 'common' config (the default) for the subcomponent 'mysubcomponent' of the subcomponent 'myapp'.

$ fab set --subcomponent "myapp.mysubcomponent" data.replicas=5 --newconfigfail

Uses the --newconfigfail switch to fail the action if a new config pair is created versus updating an existing value.
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("'set' takes one or more key=value arguments")
		}

		return Set(environment, subcomponent, args, newconfigfail)
	},
}

func init() {
	setCmd.PersistentFlags().StringVar(&environment, "environment", "common", "Environment this configuration should apply to")
	setCmd.PersistentFlags().StringVar(&subcomponent, "subcomponent", "", "Subcomponent this configuration should apply to")
	setCmd.PersistentFlags().BoolVar(&newconfigfail, "newconfigfail", false, "Fail the task if a new config pair is created versus updating an existing value")

	rootCmd.AddCommand(setCmd)
}
