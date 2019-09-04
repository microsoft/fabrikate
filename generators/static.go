package generators

import (
	"fmt"
	"io/ioutil"
	"path"

	"github.com/kyokomi/emoji"
	"github.com/microsoft/fabrikate/core"
	"github.com/microsoft/fabrikate/logger"
)

// StaticGenerator uses a static directory of resource manifests to create a rolled up multi-part manifest.
type StaticGenerator struct{}

// Generate iterates a static directory of resource manifests and creates a multi-part manifest.
func (sg *StaticGenerator) Generate(component *core.Component) (manifest string, err error) {
	logger.Info(emoji.Sprintf(":truck: Generating component '%s' statically from path %s", component.Name, component.Path))

	staticPath := component.GetStaticComponentPath()
	staticFiles, err := ioutil.ReadDir(staticPath)
	if err != nil {
		logger.Error(fmt.Sprintf("error reading from directory %s", staticPath))
		return "", err
	}

	manifests := ""
	for _, staticFile := range staticFiles {
		staticFilePath := path.Join(staticPath, staticFile.Name())

		staticFileManifest, err := ioutil.ReadFile(staticFilePath)
		if err != nil {
			return "", err
		}

		manifests += fmt.Sprintf("---\n%s\n", staticFileManifest)
	}

	return manifests, err
}

// Install is provided such that the StaticGenerator fulfills the Generator interface.
// Currently is a no-op, but could be extended to support remote static content (see #155)
func (sg *StaticGenerator) Install(component *core.Component) (err error) {
	return nil
}