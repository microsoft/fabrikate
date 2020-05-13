package commands

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/microsoft/fabrikate/pkg/encoding/yaml"
	"github.com/stretchr/testify/assert"
)

func TestSetValue(t *testing.T) {
	// This test changes the cwd. Must change back so any tests following don't break
	cwd, err := os.Getwd()
	assert.Nil(t, err)
	defer func() {
		_ = os.Chdir(cwd)
	}()

	err = os.Chdir("../testdata/set")
	assert.Nil(t, err)
	noNewConfigKeys := false

	// malformed value assignment, should return error
	err = Set("test", "", []string{"zoo"}, noNewConfigKeys, "")
	assert.NotNil(t, err)

	// malformed value assignment, should return error
	err = Set("test", "", []string{"zoo=zaa=wrong"}, noNewConfigKeys, "")
	assert.NotNil(t, err)

	// apply 'faa' as value for 'foo' in component 'test' config
	err = Set("test", "", []string{"foo=faa"}, noNewConfigKeys, "")
	assert.Nil(t, err)

	// apply 'zaa' as value for 'zoo' in subcomponent 'myapp' 'test' config
	err = Set("test", "myapp", []string{"zoo=zaa"}, noNewConfigKeys, "")
	assert.Nil(t, err)

	// create new environment
	_ = os.Remove("./config/new.yaml")
	err = Set("new", "myapp", []string{"zoo.zii=zaa"}, noNewConfigKeys, "")
	assert.Nil(t, err)

	// update deep config on existing environment
	err = Set("new", "myapp", []string{"zoo.zii=zbb"}, noNewConfigKeys, "")
	assert.Nil(t, err)

	// deep subcomponent config set
	err = Set("new", "myapp.mysubapp", []string{"foo.bar=zoo"}, noNewConfigKeys, "")
	assert.Nil(t, err)

	// deep subcomponent config set with string literal in double quotes. ex: \"k8.beta.io/load-balancer-group\"
	err = Set("new", "myservice.mysubservice", []string{"foo.bar.\"k8.beta.io/load-balancer-group\"=foo-bar-group"}, noNewConfigKeys, "")
	assert.Nil(t, err)

	err = Set("new", "myservice.mysubservice", []string{"foo.bar.line=solid"}, noNewConfigKeys, "")
	assert.Nil(t, err)

	// set existing value with new noNewConfigKeys switch on
	noNewConfigKeys = true
	err = Set("new", "myservice.mysubservice", []string{"foo.bar.\"k8.beta.io/load-balancer-group\"=foo-bar-updated"}, noNewConfigKeys, "")
	assert.Nil(t, err)

	err = Set("test", "", []string{"foo=faa"}, noNewConfigKeys, "")
	assert.Nil(t, err)

	err = Set("test", "", []string{"newfoo=faa"}, noNewConfigKeys, "")
	assert.NotNil(t, err)

	////////////////////////////////////////////////////////////////////////////////
	// Start Set from yaml file
	////////////////////////////////////////////////////////////////////////////////
	// Read target file to inject into myapp.subcomponent
	yamlFile := "inject.yaml"
	err = Set("fromfile", "myapp.mysubcomponent", []string{}, false, yamlFile)
	assert.Nil(t, err)
	bytes, err := ioutil.ReadFile(yamlFile)
	assert.Nil(t, err)

	// Parse yaml
	fromFile := map[string]interface{}{}
	err = yaml.Unmarshal(bytes, &fromFile)
	assert.Nil(t, err)

	// Read into config file
	configBytes, err := ioutil.ReadFile("config/fromfile.yaml")
	assert.Nil(t, err)
	inConfig := map[string]interface{}{}
	err = yaml.Unmarshal(configBytes, &inConfig)
	assert.Nil(t, err)

	// Config should match where myapp.mysubcomponent config == values from inject.yaml
	assert.EqualValues(t, map[string]interface{}{
		"subcomponents": map[string]interface{}{
			"myapp": map[string]interface{}{
				"subcomponents": map[string]interface{}{
					"mysubcomponent": map[string]interface{}{
						"config": fromFile,
					},
				},
			},
		},
	}, inConfig)

	err = os.Remove("config/fromfile.yaml")
	assert.Nil(t, err)
	////////////////////////////////////////////////////////////////////////////////
	// End Set from yaml file
	////////////////////////////////////////////////////////////////////////////////
}
