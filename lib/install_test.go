package lib

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInstall(t *testing.T) {

	_, err := Install("../test/fixtures/install")

	assert.Nil(t, err)
}
