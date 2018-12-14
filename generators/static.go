package generators

import (
	"fmt"
	"io/ioutil"
	"path"

	"github.com/Microsoft/fabrikate/core"
	"github.com/kyokomi/emoji"
)

func GenerateStaticComponent(component *core.Component) (manifest string, err error) {
	emoji.Printf(":truck: generating component '%s' statically from path %s\n", component.Name, component.Path)

	staticPath := path.Join(component.PhysicalPath, component.Path)
	staticFiles, err := ioutil.ReadDir(staticPath)

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
