package helm

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/kyokomi/emoji"
	"github.com/microsoft/fabrikate/internal/logger"
	"gopkg.in/yaml.v3"
)

// DependencyUpdate attempts to run `helm dependency update` on chartPath
func DependencyUpdate(chartPath string) (err error) {
	// A single helm dependency entry
	type helmDependency struct {
		Name       string
		Version    string
		Repository string
		Condition  string
	}

	// Contents of requirements.yaml or dependencies in Chart.yaml
	type helmDependencies struct {
		Dependencies []helmDependency
	}

	absChartPath := chartPath
	if isAbs := filepath.IsAbs(absChartPath); !isAbs {
		asAbs, err := filepath.Abs(chartPath)
		if err != nil {
			return err
		}
		absChartPath = asAbs
	}

	// Parse chart dependency repositories and add them if not present.
	// For both api versions v1 and v2, if requirements.yaml has dependencies
	// Chart.yaml's dependencies will be ignored.
	dependenciesYamlPath := path.Join(absChartPath, "requirements.yaml")
	if _, err := os.Stat(dependenciesYamlPath); err != nil {
		dependenciesYamlPath = path.Join(absChartPath, "Chart.yaml")
	}
	addedDepRepoList := []string{}
	if _, err := os.Stat(dependenciesYamlPath); err == nil {
		logger.Info(fmt.Sprintf("'%s' found at '%s', ensuring repositories exist on helm client", filepath.Base(dependenciesYamlPath), dependenciesYamlPath))

		bytes, err := ioutil.ReadFile(dependenciesYamlPath)
		if err != nil {
			return err
		}

		dependenciesYaml := helmDependencies{}
		if err = yaml.Unmarshal(bytes, &dependenciesYaml); err != nil {
			return err
		}

		// Add each dependency repo with a temp name
		for _, dep := range dependenciesYaml.Dependencies {
			currentRepo, _ := FindRepoNameByURL(dep.Repository)
			if currentRepo != "" {
				logger.Info(emoji.Sprintf(":pencil: Helm dependency repo already present: %v", currentRepo))
				continue
			}

			if !strings.HasPrefix(dep.Repository, "http") {
				logger.Info(emoji.Sprintf(":pencil: Skipping non-http helm dependency repo. Found '%v'", dep.Repository))
				continue
			}

			logger.Info(emoji.Sprintf(":pencil: Adding helm dependency repository '%s'", dep.Repository))
			randomUUID, err := uuid.NewRandom()
			if err != nil {
				return err
			}

			randomRepoName := randomUUID.String()
			if err := RepoAdd(randomRepoName, dep.Repository); err != nil {
				return err
			}

			addedDepRepoList = append(addedDepRepoList, randomRepoName)
		}
	}

	logger.Info(emoji.Sprintf(":helicopter: Updating helm chart's dependencies for chart in '%s'", absChartPath))
	updateCmd := exec.Command("helm", "dependency", "update", chartPath)
	var stderr bytes.Buffer
	updateCmd.Stderr = &stderr
	if err := updateCmd.Run(); err != nil {
		return fmt.Errorf("%v: %v", err, stderr.String())
	}

	// Cleanup temp dependency repositories
	for _, repo := range addedDepRepoList {
		logger.Info(emoji.Sprintf(":bomb: Removing dependency repository '%s'", repo))
		if err := RepoRemove(repo); err != nil {
			return err
		}
	}

	return err
}
