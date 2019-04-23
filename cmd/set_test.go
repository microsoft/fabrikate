package cmd

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetValue(t *testing.T) {
	err := os.Chdir("../test/fixtures/set")
	assert.Nil(t, err)
	noNewConfigKeys := false

	// malformed value assignment, should return error
	err = Set("test", "", []string{"zoo"}, noNewConfigKeys)
	assert.NotNil(t, err)

	// malformed value assignment, should return error
	err = Set("test", "", []string{"zoo=zaa=wrong"}, noNewConfigKeys)
	assert.NotNil(t, err)

	// apply 'faa' as value for 'foo' in component 'test' config
	err = Set("test", "", []string{"foo=faa"}, noNewConfigKeys)
	assert.Nil(t, err)

	// apply 'zaa' as value for 'zoo' in subcomponent 'myapp' 'test' config
	err = Set("test", "myapp", []string{"zoo=zaa"}, noNewConfigKeys)
	assert.Nil(t, err)

	// create new environment
	_ = os.Remove("./config/new.yaml")
	err = Set("new", "myapp", []string{"zoo.zii=zaa"}, noNewConfigKeys)
	assert.Nil(t, err)

	// update deep config on existing environment
	err = Set("new", "myapp", []string{"zoo.zii=zbb"}, noNewConfigKeys)
	assert.Nil(t, err)

	// deep subcomponent config set
	err = Set("new", "myapp.mysubapp", []string{"foo.bar=zoo"}, noNewConfigKeys)
	assert.Nil(t, err)

	// deep subcomponent config set with string literal in double quotes. ex: \"k8.beta.io/load-balancer-group\"
	err = Set("new", "myservice.mysubservice", []string{"foo.bar.\"k8.beta.io/load-balancer-group\"=foo-bar-group"}, noNewConfigKeys)
	assert.Nil(t, err)

	// set existing value with new noNewConfigKeys switch on
	noNewConfigKeys = true
	err = Set("test", "", []string{"foo=faa"}, noNewConfigKeys)
	assert.Nil(t, err)

	err = Set("test", "", []string{"newfoo=faa"}, noNewConfigKeys)
	assert.NotNil(t, err)

	err = os.Chdir("../../../cmd")
	assert.Nil(t, err)
}
