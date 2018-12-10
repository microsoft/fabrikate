package lib

import (
	"testing"

	"github.com/Microsoft/marina/models"
	"github.com/stretchr/testify/assert"
)

func TestIteratingDefinition(t *testing.T) {
	callbackCount := 0
	results, err := IterateComponentTree("../test/fixtures/iterator", func(path string, component *models.Component) (result string, err error) {
		callbackCount++
		return "test", nil
	})

	assert.Nil(t, err)
	assert.Equal(t, 3, len(results))
	assert.Equal(t, callbackCount, len(results))

	assert.Equal(t, results[1].PhysicalPath, "../test/fixtures/iterator/infra")
	assert.Equal(t, results[1].LogicalPath, "infra")

	assert.Equal(t, results[2].PhysicalPath, "../test/fixtures/iterator/infra/components/efk")
	assert.Equal(t, results[2].LogicalPath, "infra/efk")
}
