package generatable

import (
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type Static struct {
	ComponentPath []string // list representation of the component structure; used to generate the generation path
	ManifestPath  string   // path to static manifests
}

func (s Static) Validate() error {
	if _, err := os.Stat(s.ManifestPath); os.IsNotExist(err) {
		return err
	}

	return nil
}

func (s Static) Generate() error {
	files, err := ioutil.ReadDir(s.ManifestPath)
	if err != nil {
		return err
	}

	// Load all yaml manifests
	var manifests []string
	for _, file := range files {
		if !file.IsDir() {
			ext := strings.ToLower(filepath.Ext(file.Name()))
			if ext == ".yaml" || ext == ".yml" {
				manifestFilePath := path.Join(file.Name())
				content, err := ioutil.ReadFile(manifestFilePath)
				if err != nil {
					return err
				}
				manifests = append(manifests, string(content))
			}
		}
	}

	// delete existing generation path
	generatePath := s.GetGeneratePath()
	if err := os.Remove(generatePath); err != nil {
		return err
	}

	// Write manifests to generation path
	unifiedManifest := strings.Join(manifests, "\n---\n")
	if err := ioutil.WriteFile(generatePath, []byte(unifiedManifest), 0755); err != nil {
		return err
	}

	return nil
}

func (s Static) GetGeneratePath() string {
	componentName := strings.Join(s.ComponentPath, "__")
	pathSegments := []string{generateDirName, componentName}

	return path.Join(pathSegments...) + ".yaml"
}
