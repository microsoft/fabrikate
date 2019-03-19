package cmd

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetValue(t *testing.T) {
	os.Chdir("../test/fixtures/set")

	// malformed value assignment, should return error
	err := Set("test", "", []string{"zoo"})
	assert.NotNil(t, err)

	// malformed value assignment, should return error
	err = Set("test", "", []string{"zoo=zaa=wrong"})
	assert.NotNil(t, err)

	// apply 'faa' as value for 'foo' in component 'test' config
	err = Set("test", "", []string{"foo=faa"})
	assert.Nil(t, err)

	// apply 'zaa' as value for 'zoo' in subcomponent 'myapp' 'test' config
	err = Set("test", "myapp", []string{"zoo=zaa"})
	assert.Nil(t, err)

	// create new environment
	err = os.Remove("./config/new.yaml")
	assert.Nil(t, err)
	err = Set("new", "myapp", []string{"zoo.zii=zaa"})
	assert.Nil(t, err)

	// update deep config on existing environment
	err = Set("new", "myapp", []string{"zoo.zii=zbb"})
	assert.Nil(t, err)

	// deep subcomponent config set
	err = Set("new", "myapp.mysubapp", []string{"foo.bar=zoo"})
	assert.Nil(t, err)

	os.Chdir("../../../cmd")
}
