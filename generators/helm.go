package generators

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/kyokomi/emoji"
	"github.com/microsoft/fabrikate/core"
	"github.com/otiai10/copy"
	log "github.com/sirupsen/logrus"
	"github.com/timfpark/yaml"
)

// HelmGenerator provides 'helm generate' generator functionality to Fabrikate
type HelmGenerator struct {
}

func addNamespaceToManifests(manifests, namespace string) (namespacedManifests string, err error) {
	splitManifest := strings.Split(manifests, "\n---")

	for _, manifest := range splitManifest {
		parsedManifest := make(map[interface{}]interface{})
		if err := yaml.Unmarshal([]byte(manifest), &parsedManifest); err != nil {
			return "", err
		}

		// strip any empty entries
		if len(parsedManifest) == 0 {
			continue
		}

		if parsedManifest["metadata"] != nil {
			metadataMap := parsedManifest["metadata"].(map[interface{}]interface{})
			if metadataMap["namespace"] == nil {
				metadataMap["namespace"] = namespace
			}
		}

		updatedManifest, err := yaml.Marshal(&parsedManifest)
		if err != nil {
			return "", err
		}

		namespacedManifests += fmt.Sprintf("---\n%s\n", updatedManifest)
	}

	return namespacedManifests, nil
}

// makeHelmRepoPath returns the path where the components helm charts are
// located -- will be an entire helm repo if `method: git` or just the target
// chart if `method: helm`
func (hg *HelmGenerator) makeHelmRepoPath(c *core.Component) string {
	// `method: git` will clone the entire helm repo; uses path to point to chart dir
	if c.Method == "git" || c.Method == "helm" {
		return path.Join(c.PhysicalPath, "helm_repos", c.Name)
	}

	return c.PhysicalPath
}

// getChartPath returns the absolute path to the directory containing the
// Chart.yaml
func (hg *HelmGenerator) getChartPath(c *core.Component) (string, error) {
	if c.Method == "helm" || c.Method == "git" {
		absHelmPath, err := filepath.Abs(hg.makeHelmRepoPath(c))
		if err != nil {
			return "", err
		}
		switch c.Method {
		case "git":
			// method: git downloads the entire repo into _helm_chart and the dir containing Chart.yaml specified by Path
			return path.Join(absHelmPath, c.Path), nil
		case "helm":
			// method: helm only downloads target chart into _helm_chart
			return absHelmPath, nil
		}
	}

	// Default to `method: local` and use the Path provided as location of the chart
	return filepath.Abs(path.Join(c.PhysicalPath, c.Path))
}

// Generate returns the helm templated manifests specified by this component.
func (hg *HelmGenerator) Generate(component *core.Component) (manifest string, err error) {
	log.Info(emoji.Sprintf(":truck: Generating component '%s' with helm with repo %s", component.Name, component.Source))

	configYaml, err := yaml.Marshal(&component.Config.Config)
	if err != nil {
		log.Errorf("Marshalling config yaml for helm generated component '%s' failed with: %s\n", component.Name, err.Error())
		return "", err
	}

	// Write helm config to temporary file in tmp folder
	randomString, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}
	overriddenValuesFileName := fmt.Sprintf("%s.yaml", randomString.String())
	absOverriddenPath := path.Join(os.TempDir(), overriddenValuesFileName)
	log.Debug(emoji.Sprintf(":pencil: Writing config %s to %s\n", configYaml, absOverriddenPath))
	if err = ioutil.WriteFile(absOverriddenPath, configYaml, 0777); err != nil {
		return "", err
	}

	// Default to `default` namespace unless provided
	namespace := "default"
	if component.Config.Namespace != "" {
		namespace = component.Config.Namespace
	}

	// Run `helm template` on the chart using the config stored in temp dir
	chartPath, err := hg.getChartPath(component)
	if err != nil {
		return "", err
	}
	log.Infof("Runing `helm template` on template '%s'\n", chartPath)
	output, err := exec.Command("helm", "template", chartPath, "--values", absOverriddenPath, "--name", component.Name, "--namespace", namespace).CombinedOutput()
	if err != nil {
		log.Errorf("helm template failed with:\n%s: %s", err, output)
		return "", err
	}
	stringManifests := string(output)

	// helm template does not inject namespace unless chart directly provides support for it: https://github.com/helm/helm/issues/3553
	// some helm templates expect Tiller to inject namespace, so enable Fabrikate component designer to
	// opt into injecting these namespaces manually.  We should reassess if this is necessary after Helm 3 is released and client side
	// templating really becomes a first class function in Helm.
	if component.Config.InjectNamespace && component.Config.Namespace != "" {
		stringManifests, err = addNamespaceToManifests(stringManifests, component.Config.Namespace)
	}

	return stringManifests, err
}

// Install installs the helm chart specified by the passed component and performs any
// helm lifecycle events needed.
func (hg *HelmGenerator) Install(c *core.Component) (err error) {
	// Install the chart
	if (c.Method == "helm" || c.Method == "git") && c.Source != "" && c.Path != "" {
		// Download the helm chart
		helmRepoPath := hg.makeHelmRepoPath(c)
		switch c.Method {
		case "helm":
			log.Info(emoji.Sprintf(":helicopter: Component '%s' requesting helm chart '%s' from helm repository '%s'", c.Name, c.Path, c.Source))
			if err = hd.downloadChart(c.Source, c.Path, helmRepoPath); err != nil {
				return err
			}
		case "git":
			// Clone whole repo into helm repo path
			log.Info(emoji.Sprintf(":helicopter: Component '%s' requesting helm chart in path '%s' from git repository '%s'", c.Name, c.Source, c.PhysicalPath))
			if err = core.CloneRepo(c.Source, c.Version, helmRepoPath, c.Branch); err != nil {
				return err
			}
		}
	}

	// Update chart dependencies -- don't fail if error is returned, but throw warning
	chartPath, err := hg.getChartPath(c)
	if err != nil {
		return err
	}
	log.Info(emoji.Sprintf(":helicopter: Updating helm chart's dependencies for component '%s'", c.Name))
	if output, err := exec.Command("helm", "dependency", "update", chartPath).CombinedOutput(); err != nil {
		log.Warn(emoji.Sprintf(":no_entry_sign: Updating chart dependencies failed for chart '%s' in component '%s'; run `helm dependency update %s` for more error details.\n%s: %s", c.Name, c.Path, chartPath, err, output))
	}

	return err
}

// helmDownloader is a thread safe chart downloader which can add/remove
// helm repositories in a thread safe way.
// Shelling `helm repo add` is not thread safe. If 2 callers do it at the same
// time, only one will go through as they will both read in existing helm repo
// list at the same time, modify, and the write out new ones; Making one get
// overwritten.
type helmDownloader struct {
	mu sync.RWMutex
}

var hd = helmDownloader{}

// downloadChart downloads a target `chart` from `repo` and places it in `into`
// -- `into` will be the dir containing Chart.yaml
// The function will add a temporary helm repo, fetch from it, and then remove
// the temporary repo. This is a to get around a limitation in Helm 2.
// see: https://github.com/helm/helm/issues/4527
func (hd *helmDownloader) downloadChart(repo, chart, into string) (err error) {
	// generate random name to store repo in helm in temporarily
	randomUUID, err := uuid.NewRandom()
	if err != nil {
		return err
	}
	randomName := randomUUID.String()
	log.Infof("Adding temporary helm repo %s => %s", repo, randomName)
	hd.mu.Lock()
	if output, err := exec.Command("helm", "repo", "add", randomName, repo).CombinedOutput(); err != nil {
		hd.mu.Unlock()
		log.Error(emoji.Sprintf(":no_entry_sign: Failed adding helm repository '%s'\n%s: %s", repo, err, output))
		return err
	}
	hd.mu.Unlock()

	// Fetch chart to random temp dir
	chartName := fmt.Sprintf("%s/%s", randomName, chart)
	randomDir := path.Join(os.TempDir(), randomName)
	log.Info(emoji.Sprintf(":helicopter: Fetching helm chart '%s' into '%s'", chart, randomDir))
	if output, err := exec.Command("helm", "fetch", "--untar", "--untardir", randomDir, chartName).CombinedOutput(); err != nil {
		log.Error(emoji.Sprintf(":no_entry_sign: Failed fetching helm chart '%s' from repo '%s'\n%s: %s", chart, repo, err, output))
		return err
	}

	// Remove repository once completed
	log.Info(emoji.Sprintf(":bomb: Removing temporary helm repo %s", randomName))
	hd.mu.Lock()
	if output, err := exec.Command("helm", "repo", "remove", randomName).CombinedOutput(); err != nil {
		hd.mu.Unlock()
		log.Error(emoji.Sprintf(":no_entry_sign: Failed to `helm repo remove %s`\n%s: %s", randomName, err, output))
	}
	hd.mu.Unlock()

	// copy chart to target `into` dir
	chartDirectoryInRandomDir := path.Join(randomDir, chart)
	return copy.Copy(chartDirectoryInRandomDir, into)
}
