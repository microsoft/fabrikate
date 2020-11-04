package generators

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCleanK8sManifest(t *testing.T) {
	manifest := `
---
this should be removed
---
this: is a valid map and should stay
another:
  entry: in the map
---
this should be removed as well
---
# This should be removed
---
---
this is another: valid map
should: not be removed
---
# Another to be removed
`
	cleaned, err := cleanK8sManifest(manifest)
	assert.Nil(t, err)
	entries := strings.Split(cleaned, "\n---")
	assert.Equal(t, 2, len(entries))
}
