package generators

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"reflect"
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

type namespaceInjectionResponse struct {
	namespacedManifest *[]byte
	err                error
	warn               *string
}

// func addNamespaceToManifests(manifests, namespace string) (namespacedManifests string, err error) {
func addNamespaceToManifests(manifests, namespace string) chan namespaceInjectionResponse {
	respChan := make(chan namespaceInjectionResponse)
	syncGroup := sync.WaitGroup{}
	splitManifest := strings.Split(manifests, "\n---")

	// Wait for all manifests to be iterated over then close the channel
	syncGroup.Add(len(splitManifest))
	go func() {
		syncGroup.Wait()
		close(respChan)
	}()

	// Iterate over all manifests, decrementing the wait group for every channel put
	for _, manifest := range splitManifest {
		go func(manifest string) {
			parsedManifest := make(map[interface{}]interface{})

			// Push a warning if unable to unmarshal
			if err := yaml.Unmarshal([]byte(manifest), &parsedManifest); err != nil {
				warning := emoji.Sprintf(":question: Unable to unmarshal manifest into type '%s', this is most likely a warning message outputted from `helm template`. Skipping namespace injection of '%s' into manifest: '%s'", reflect.TypeOf(parsedManifest), namespace, manifest)
				respChan <- namespaceInjectionResponse{warn: &warning}
				syncGroup.Done()
				return
			}

			// strip any empty entries
			if len(parsedManifest) == 0 {
				syncGroup.Done()
				return
			}

			// Inject the namespace
			if parsedManifest["metadata"] != nil {
				metadataMap := parsedManifest["metadata"].(map[interface{}]interface{})
				if metadataMap["namespace"] == nil {
					metadataMap["namespace"] = namespace
				}
			}

			// Marshal updated manifest and put the response on channel
			updatedManifest, err := yaml.Marshal(&parsedManifest)
			if err != nil {
				respChan <- namespaceInjectionResponse{err: err}
				syncGroup.Done()
				return
			}
			respChan <- namespaceInjectionResponse{namespacedManifest: &updatedManifest}
			syncGroup.Done()
		}(manifest)
	}

	return respChan
}

// cleanK8sManifest attempts to remove any invalid entries in k8s yaml.
// If any entries after being split by "---" are not a map or are empty, they are removed
func cleanK8sManifest(manifests string) (cleanedManifests string, err error) {
	splitManifest := strings.Split(manifests, "\n---")

	for _, manifest := range splitManifest {
		parsedManifest := make(map[interface{}]interface{})

		// Log a warning if unable to unmarshal; skip the entry
		if err := yaml.Unmarshal([]byte(manifest), &parsedManifest); err != nil {
			warning := emoji.Sprintf(":question: Unable to unmarshal manifest into type '%s', this is most likely a warning message outputted from `helm template`.\nRemoving manifest entry: '%s'\nUnmarshal error encountered: '%s'", reflect.TypeOf(parsedManifest), manifest, err)
			log.Warn(warning)
			continue
		}

		// Remove empty entries
		if len(parsedManifest) == 0 {
			continue
		}

		cleanedManifests += fmt.Sprintf("---\n%s\n", manifest)
	}

	return cleanedManifests, err
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
	log.Info(emoji.Sprintf(":memo: Running `helm template` on template '%s'", chartPath))
	output, err := exec.Command("helm", "template", chartPath, "--values", absOverriddenPath, "--name", component.Name, "--namespace", namespace).CombinedOutput()
	if err != nil {
		log.Errorf("helm template failed with:\n%s: %s", err, output)
		return "", err
	}
	// Remove any empty/non-map entries in manifests
	log.Info(emoji.Sprintf(":scissors: Removing empty entries from generated manifests from chart '%s'", chartPath))
	stringManifests, err := cleanK8sManifest(string(output))
	if err != nil {
		return "", err
	}

	// helm template does not inject namespace unless chart directly provides support for it: https://github.com/helm/helm/issues/3553
	// some helm templates expect Tiller to inject namespace, so enable Fabrikate component designer to
	// opt into injecting these namespaces manually.  We should reassess if this is necessary after Helm 3 is released and client side
	// templating really becomes a first class function in Helm.
	if component.Config.InjectNamespace && component.Config.Namespace != "" {
		log.Info(emoji.Sprintf(":syringe: Injecting namespace '%s' into manifests for component '%s'", component.Config.Namespace, component.Name))
		namespacedManifests := ""
		for resp := range addNamespaceToManifests(stringManifests, component.Config.Namespace) {
			// If error; return the error immediately
			if resp.err != nil {
				log.Error(emoji.Sprintf(":exclamation: Encountered error while injecting namespace '%s' into manifests for component '%s':\n%s", component.Config.Namespace, component.Name, resp.err))
				return stringManifests, resp.err
			}

			// If warning; just log the warning
			if resp.warn != nil {
				log.Warn(emoji.Sprintf(":question: Encountered warning while injecting namespace '%s' into manifests for component '%s':\n%s", component.Config.Namespace, component.Name, *resp.warn))
			}

			// Add the manifest if one was returned
			if resp.namespacedManifest != nil {
				namespacedManifests += fmt.Sprintf("---\n%s\n", *resp.namespacedManifest)
			}
		}
		stringManifests = namespacedManifests
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
			// Update chart dependencies in chart path -- this is manually done here but automatically done in downloadChart in the case of `method: helm`
			chartPath, err := hg.getChartPath(c)
			if err != nil {
				return err
			}
			if err = updateHelmChartDep(chartPath); err != nil {
				return err
			}
		}
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
	if err = copy.Copy(chartDirectoryInRandomDir, into); err != nil {
		return err
	}

	// update/fetch dependencies for the chart
	return updateHelmChartDep(into)
}

// updateHelmChartDep attempts to run `helm dependency update` on chartPath
func updateHelmChartDep(chartPath string) (err error) {
	// A single helm dependency entry
	type helmDependency struct {
		Name       string
		Version    string
		Repository string
		Condition  string
	}

	// Contents of requirements.yaml for a helm chart
	type helmRequirements struct {
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

	// Parse chart dependency repositories and add them if not present
	requirementsYamlPath := path.Join(absChartPath, "requirements.yaml")
	addedDepRepoList := []string{}
	if _, err := os.Stat(requirementsYamlPath); err == nil {
		log.Infof("requirements.yaml found at '%s', ensuring repositories exist on helm client", requirementsYamlPath)
		bytes, err := ioutil.ReadFile(requirementsYamlPath)
		if err != nil {
			return err
		}
		requirementsYaml := helmRequirements{}
		if err = yaml.Unmarshal(bytes, &requirementsYaml); err != nil {
			return err
		}

		// Add each dependency repo with a temp name
		for _, dep := range requirementsYaml.Dependencies {
			log.Info(emoji.Sprintf(":pencil: Adding helm dependency repository '%s'", dep.Repository))
			randomUUID, err := uuid.NewRandom()
			if err != nil {
				return err
			}
			randomRepoName := randomUUID.String()
			hd.mu.Lock()
			if output, err := exec.Command("helm", "repo", "add", randomRepoName, dep.Repository).CombinedOutput(); err != nil {
				hd.mu.Unlock()
				log.Error(emoji.Sprintf(":no_entry_sign: Failed to add helm dependency repository '%s' for chart '%s':\n%s", dep.Repository, chartPath, output))
				return err
			}
			hd.mu.Unlock()
			addedDepRepoList = append(addedDepRepoList, randomRepoName)
		}
	}

	// Update dependencies
	log.Info(emoji.Sprintf(":helicopter: Updating helm chart's dependencies for chart in '%s'", absChartPath))
	if output, err := exec.Command("helm", "dependency", "update", chartPath).CombinedOutput(); err != nil {
		log.Warn(emoji.Sprintf(":no_entry_sign: Updating chart dependencies failed for chart in '%s'; run `helm dependency update %s` for more error details.\n%s: %s", absChartPath, absChartPath, err, output))
		return err
	}

	// Cleanup temp dependency repositories
	for _, repo := range addedDepRepoList {
		log.Info(emoji.Sprintf(":bomb: Removing dependency repository '%s'", repo))
		hd.mu.Lock()
		if output, err := exec.Command("helm", "repo", "remove", repo).CombinedOutput(); err != nil {
			hd.mu.Unlock()
			log.Error(output)
			return err
		}
		hd.mu.Unlock()
	}

	return err
}
