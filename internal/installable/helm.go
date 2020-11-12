package installable

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/microsoft/fabrikate/internal/helm"
	"github.com/microsoft/fabrikate/internal/url"
)

type Helm struct {
	URL     string
	Chart   string
	Version string
}

func (h Helm) Install() error {
	// Pull to a temporary directory
	tmpHelmDir, err := ioutil.TempDir("", "fabrikate")
	defer os.RemoveAll(tmpHelmDir)
	if err != nil {
		return err
	}
	if err := helm.Pull(h.URL, h.Chart, h.Version, tmpHelmDir); err != nil {
		return err
	}

	componentPath, err := h.GetInstallPath()
	if err != nil {
		return err
	}
	if err := os.RemoveAll(componentPath); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(componentPath), 0755); err != nil {
		return err
	}

	// Move the extracted chart from tmp to the _component dir
	extractedChartPath := path.Join(tmpHelmDir, h.Chart)
	if err := os.Rename(extractedChartPath, componentPath); err != nil {
		return err
	}

	return nil
}

func (h Helm) GetInstallPath() (string, error) {
	urlPath, err := url.ToPath(h.URL)
	if err != nil {
		return "", err
	}
	var version string
	if h.Version != "" {
		version = h.Version
	} else {
		version = "latest"
	}

	componentPath := path.Join(installDirName, urlPath, h.Chart, version)
	return componentPath, nil
}

func (h Helm) Validate() error {
	if h.URL == "" {
		return fmt.Errorf(`URL must be non-zero length, "%v" provided`, h.URL)
	}
	if h.Chart == "" {
		return fmt.Errorf(`Chart must be non-zero length, "%v" provided`, h.Chart)
	}

	return nil
}
