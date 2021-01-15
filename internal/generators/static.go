package generators

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/kyokomi/emoji"
	"github.com/microsoft/fabrikate/internal/core"
	"github.com/microsoft/fabrikate/internal/logger"
)

// StaticGenerator uses a static directory of resource manifests to create a rolled up multi-part manifest.
type StaticGenerator struct{}

// GetStaticManifestsPath returns the path where a static components YAML
// manifest files are located, defaulting `path` unless `method: http` in which
// case `<path>/components/<name>` is used.
func GetStaticManifestsPath(c core.Component) (string, error) {
	// if strings.EqualFold(c.Method, "http") {
	// 	return path.Join(c.PhysicalPath, "components", c.Name)
	// }

	// return path.Join(c.PhysicalPath, c.Path)

	i, err := c.ToInstallable()
	if err != nil {
		return "", err
	}
	installPath, err := i.GetInstallPath()
	if err != nil {
		return "", err
	}

	return installPath, nil
}

// Generate iterates a static directory of resource manifests and creates a multi-part manifest.
func (sg *StaticGenerator) Generate(component *core.Component) (manifest string, err error) {
	logger.Info(emoji.Sprintf(":truck: Generating component '%s' statically from path %s", component.Name, component.Path))

	staticPath, err := GetStaticManifestsPath(*component)
	if err != nil {
		return "", err
	}
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

// Install for StaticGenerator gives the ability to point to a single yaml
// manifest over `method: http`; This is a noop for any all other methods.
func (sg *StaticGenerator) Install(c *core.Component) (err error) {
	if strings.EqualFold(c.Method, "http") {
		// validate that `Source` points to a yaml file
		validSourceExtensions := []string{".yaml", ".yml"}
		isValidExtension := func() bool {
			sourceExt := strings.ToLower(filepath.Ext(c.Source))
			for _, validExt := range validSourceExtensions {
				if sourceExt == validExt {
					return true
				}
			}
			return false
		}()
		if !isValidExtension {
			return fmt.Errorf("source for 'static' component '%s' must end in one of %v; given: '%s'", c.Name, validSourceExtensions, c.Source)
		}

		response, err := http.Get(c.Source)
		if err != nil {
			return err
		}
		defer response.Body.Close()

		componentsPath := path.Join(c.PhysicalPath, "components", c.Name)
		if err := os.MkdirAll(componentsPath, 0777); err != nil {
			return err
		}

		// Write the downloaded resource manifest file
		out, err := os.Create(path.Join(componentsPath, c.Name+".yaml"))
		if err != nil {
			logger.Error(emoji.Sprintf(":no_entry_sign: Error occurred in install for component '%s'\nError: %s", c.Name, err))
			return err
		}
		defer out.Close()

		if _, err = io.Copy(out, response.Body); err != nil {
			logger.Error(emoji.Sprintf(":no_entry_sign: Error occurred in writing manifest file for component '%s'\nError: %s", c.Name, err))
			return err
		}
	}

	return nil
}
