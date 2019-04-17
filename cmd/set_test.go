package cmd

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetValue(t *testing.T) {
	err := os.Chdir("../test/fixtures/set")
	assert.Nil(t, err)
	newconfigfail := false

	// malformed value assignment, should return error
	err = Set("test", "", []string{"zoo"}, newconfigfail)
	assert.NotNil(t, err)

	// malformed value assignment, should return error
	err = Set("test", "", []string{"zoo=zaa=wrong"}, newconfigfail)
	assert.NotNil(t, err)

	// apply 'faa' as value for 'foo' in component 'test' config
	err = Set("test", "", []string{"foo=faa"}, newconfigfail)
	assert.Nil(t, err)

	// apply 'zaa' as value for 'zoo' in subcomponent 'myapp' 'test' config
	err = Set("test", "myapp", []string{"zoo=zaa"}, newconfigfail)
	assert.Nil(t, err)

	// create new environment
	_ = os.Remove("./config/new.yaml")
	err = Set("new", "myapp", []string{"zoo.zii=zaa"}, newconfigfail)
	assert.Nil(t, err)

	// update deep config on existing environment
	err = Set("new", "myapp", []string{"zoo.zii=zbb"}, newconfigfail)
	assert.Nil(t, err)

	// deep subcomponent config set
	err = Set("new", "myapp.mysubapp", []string{"foo.bar=zoo"}, newconfigfail)
	assert.Nil(t, err)

	// set existing value with new newconfigfail switch on
	newconfigfail = true
	err = Set("test", "", []string{"foo=faa"}, newconfigfail)
	assert.Nil(t, err)

	err = Set("test", "", []string{"newfoo=faa"}, newconfigfail)
	assert.NotNil(t, err)

	err = os.Chdir("../../../cmd")
	assert.Nil(t, err)
}
