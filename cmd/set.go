package cmd

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io/ioutil"
	"reflect"
	"strings"

	"github.com/microsoft/fabrikate/core"
	"github.com/microsoft/fabrikate/util"
	"github.com/spf13/cobra"
	"github.com/timfpark/yaml"
)

// SplitPathValuePairs splits array of key/value pairs and returns array of path value pairs ([] PathValuePair) //
func SplitPathValuePairs(pathValuePairStrings []string) (pathValuePairs []core.PathValuePair, err error) {
	for _, pathValuePairString := range pathValuePairStrings {
		pathValuePairParts := strings.Split(pathValuePairString, "=")

		errMessage := "%s is not a properly formated configuration key/value pair"

		if len(pathValuePairParts) != 2 {
			return pathValuePairs, fmt.Errorf(errMessage, pathValuePairString)
		}

		pathParts, err := SplitPathParts(pathValuePairParts[0])

		if err != nil {
			return pathValuePairs, fmt.Errorf(errMessage, pathValuePairString)
		}

		pathValuePair := core.PathValuePair{
			Path:  pathParts,
			Value: pathValuePairParts[1],
		}

		pathValuePairs = append(pathValuePairs, pathValuePair)
	}

	return pathValuePairs, nil
}

// SplitPathParts splits path string at . while ignoring string literals enclosed in quotes (".") and returns an array //
func SplitPathParts(path string) (pathParts []string, err error) {

	csv := csv.NewReader(strings.NewReader(path))

	// Comma is the field delimiter. Dot (.) will be the value for config key
	csv.Comma = '.'

	// setting it to true, a quote may appear in an unquoted field and a non-doubled quote may appear in a quoted field.
	csv.LazyQuotes = true

	// FieldsPerRecord is the number of expected fields per record.
	// > 0: Read requires each record to have the given number of fields.
	// == 0, Read sets it to the number of fields in the first record, so that future records must have the same field count.
	// < 0, no check is made and config key may have a variable number of fields.
	csv.FieldsPerRecord = -1

	// Read parts and the error
	parts, err := csv.Read()

	// return err and empty parts
	if err != nil {
		return nil, err
	}

	// return key parts
	return parts, nil
}

// Set implements the 'set' command. It takes an environment, a set of config path / value strings (and a subcomponent if the config
// should be set on a subcomponent versus the component itself) and sets the config in the appropriate config file,
// writing the result out to disk at the end.
func Set(environment string, subcomponent string, pathValuePairStrings []string, noNewConfigKeys bool, inputFile string) (err error) {

	subcomponentPath := []string{}
	if len(subcomponent) > 0 {
		subcomponentPath = strings.Split(subcomponent, ".")
	}

	componentConfig := core.NewComponentConfig(".")

	// Load input file if provided
	inputFileValuePairList := []string{}
	if inputFile != "" {
		bytes, err := ioutil.ReadFile(inputFile)
		if err != nil {
			return err
		}
		yamlContent := map[string]interface{}{}
		*yaml.DefaultMapType = reflect.TypeOf(map[string]interface{}{})
		err = yaml.Unmarshal(bytes, &yamlContent)
		if err != nil {
			return err
		}

		// Flatten the map
		flattenedInputFileContentMap := util.FlattenMap(yamlContent, ".", []string{})

		// Append all key/value in map to the flattened list
		for k, v := range flattenedInputFileContentMap {
			// Join to PathValue strings with "="
			valueAsString := fmt.Sprintf("%v", v)
			joined := strings.Join([]string{k, valueAsString}, "=")
			inputFileValuePairList = append(inputFileValuePairList, joined)
		}
	}

	pathValuePairs, err := SplitPathValuePairs(append(inputFileValuePairList, pathValuePairStrings...))

	if err != nil {
		return err
	}

	if err := componentConfig.Load(environment); err != nil {
		return err
	}

	newConfigError := errors.New("new configuration was specified and the --no-new-config-keys switch is on")

	for _, pathValue := range pathValuePairs {
		if noNewConfigKeys {
			if !componentConfig.HasSubcomponentConfig(subcomponentPath) {
				return newConfigError
			}

			sc := componentConfig.GetSubcomponentConfig(subcomponentPath)

			if !sc.HasComponentConfig(pathValue.Path) {
				return newConfigError
			}
		}

		componentConfig.SetConfig(subcomponentPath, pathValue.Path, pathValue.Value)
	}

	return componentConfig.Write(environment)
}

var subcomponent string
var environment string
var noNewConfigKeys bool
var inputFile string

var setCmd = &cobra.Command{
	Use:   "set <config> [--subcomponent subcomponent] [--file <my-yaml-file.yaml>] <path1>=<value1> <path2>=<value2> ...",
	Short: "Sets a config value for a component for a particular config environment in the Fabrikate definition.",
	Long: `Sets a config value for a component for a particular config environment in the Fabrikate definition.
eg.
$ fab set --environment prod data.replicas=4 username="ops"

Sets the value of 'data.replicas' equal to 4 and 'username' equal to 'ops' in the 'prod' config for the current component.

$ fab set --subcomponent "myapp" endpoint="east-db" 

Sets the value of 'endpoint' equal to 'east-db' in the 'common' config (the default) for subcomponent 'myapp'.

$ fab set --subcomponent "myapp.mysubcomponent" data.replicas=5 

Sets the subkey "replicas" in the key 'data' equal to 5 in the 'common' config (the default) for the subcomponent 'mysubcomponent' of the subcomponent 'myapp'.

$ fab set --subcomponent "myapp.mysubcomponent" data.replicas=5 --no-new-config-keys

Use the --no-new-config-keys switch to prevent the creation of new config.
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 && inputFile == "" {
			return errors.New("'set' takes one or more key=value arguments and/or a --file")
		}

		return Set(environment, subcomponent, args, noNewConfigKeys, inputFile)
	},
}

func init() {
	setCmd.PersistentFlags().StringVar(&environment, "environment", "common", "Environment this configuration should apply to")
	setCmd.PersistentFlags().StringVar(&subcomponent, "subcomponent", "", "Subcomponent this configuration should apply to")
	setCmd.PersistentFlags().BoolVar(&noNewConfigKeys, "no-new-config-keys", false, "'Prevent creation of new config keys and only allow updating existing config values.")
	setCmd.Flags().StringVarP(&inputFile, "file", "f", "", "Path to a single YAML file which can be read in and the values of which will be set; note '.' can not occur in keys and list values are not supported.")

	rootCmd.AddCommand(setCmd)
}
