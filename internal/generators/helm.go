package generators

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/kyokomi/emoji"
	"github.com/microsoft/fabrikate/internal/core"
	"github.com/microsoft/fabrikate/internal/git"
	"github.com/microsoft/fabrikate/internal/helm"
	"github.com/microsoft/fabrikate/internal/logger"
	"github.com/timfpark/yaml"
)

// HelmGenerator provides 'helm generate' generator functionality to Fabrikate
type HelmGenerator struct{}

type namespaceInjectionResponse struct {
	index              int
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
	for index, manifest := range splitManifest {
		go func(index int, manifest string) {
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
			respChan <- namespaceInjectionResponse{index: index, namespacedManifest: &updatedManifest}
			syncGroup.Done()
		}(index, manifest)
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
			logger.Warn(warning)
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
	installer, err := c.ToInstallable()
	if err != nil {
		return "", err
	}
	installPath, err := installer.GetInstallPath()
	if err != nil {
		return "", err
	}
	return installPath, nil
	// if c.Method == "helm" || c.Method == "git" {
	// 	absHelmPath, err := filepath.Abs(hg.makeHelmRepoPath(c))
	// 	if err != nil {
	// 		return "", err
	// 	}
	// 	switch c.Method {
	// 	case "git":
	// 		// method: git downloads the entire repo into _helm_chart and the dir containing Chart.yaml specified by Path
	// 		return path.Join(absHelmPath, c.Path), nil
	// 	case "helm":
	// 		// method: helm only downloads target chart into _helm_chart
	// 		return absHelmPath, nil
	// 	}
	// }

	// // Default to `method: local` and use the Path provided as location of the chart
	// return filepath.Abs(path.Join(c.PhysicalPath, c.Path))
}

// Generate returns the helm templated manifests specified by this component.
func (hg *HelmGenerator) Generate(component *core.Component) (manifest string, err error) {
	logger.Info(emoji.Sprintf(":truck: Generating component '%s' with helm with repo %s", component.Name, component.Source))

	configYaml, err := yaml.Marshal(&component.Config.Config)
	if err != nil {
		logger.Error(fmt.Sprintf("Marshalling config yaml for helm generated component '%s' failed with: %s\n", component.Name, err.Error()))
		return "", err
	}

	// Write helm config to temporary file in tmp folder
	randomString, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}
	overriddenValuesFileName := fmt.Sprintf("%s.yaml", randomString.String())
	absOverriddenPath := path.Join(os.TempDir(), overriddenValuesFileName)
	defer os.Remove(absOverriddenPath)

	logger.Debug(emoji.Sprintf(":pencil: Writing config %s to %s\n", configYaml, absOverriddenPath))
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
	logger.Info(emoji.Sprintf(":memo: Running `helm template` on template '%s'", chartPath))
	output, err := exec.Command("helm", "template", component.Name, chartPath, "--values", absOverriddenPath, "--namespace", namespace).CombinedOutput()
	if err != nil {
		logger.Error(fmt.Sprintf("helm template failed with:\n%s: %s", err, output))
		return "", err
	}
	// Remove any empty/non-map entries in manifests
	logger.Info(emoji.Sprintf(":scissors: Removing empty entries from generated manifests from chart '%s'", chartPath))
	stringManifests, err := cleanK8sManifest(string(output))
	if err != nil {
		return "", err
	}

	// helm template does not inject namespace unless chart directly provides support for it: https://github.com/helm/helm/issues/3553
	// some helm templates expect Tiller to inject namespace, so enable Fabrikate component designer to
	// opt into injecting these namespaces manually.  We should reassess if this is necessary after Helm 3 is released and client side
	// templating really becomes a first class function in Helm.
	if component.Config.InjectNamespace && component.Config.Namespace != "" {
		logger.Info(emoji.Sprintf(":syringe: Injecting namespace '%s' into manifests for component '%s'", component.Config.Namespace, component.Name))
		var successes []namespaceInjectionResponse
		for resp := range addNamespaceToManifests(stringManifests, component.Config.Namespace) {
			// If error; return the error immediately
			if resp.err != nil {
				logger.Error(emoji.Sprintf(":exclamation: Encountered error while injecting namespace '%s' into manifests for component '%s':\n%s", component.Config.Namespace, component.Name, resp.err))
				return stringManifests, resp.err
			}

			// If warning; just log the warning
			if resp.warn != nil {
				logger.Warn(emoji.Sprintf(":question: Encountered warning while injecting namespace '%s' into manifests for component '%s':\n%s", component.Config.Namespace, component.Name, *resp.warn))
			}

			// Add the manifest if one was returned
			if resp.namespacedManifest != nil {
				successes = append(successes, resp)
			}
		}

		sort.Slice(successes, func(i, j int) bool {
			return successes[i].index < successes[j].index
		})

		namespacedManifests := ""
		for _, resp := range successes {
			namespacedManifests += fmt.Sprintf("---\n%s\n", *resp.namespacedManifest)
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
			logger.Info(emoji.Sprintf(":helicopter: Component '%s' requesting helm chart '%s' from helm repository '%s'", c.Name, c.Path, c.Source))
			// Pull to a temporary directory
			tmpHelmDir, err := ioutil.TempDir("", "fabrikate")
			defer os.RemoveAll(tmpHelmDir)
			if err != nil {
				return err
			}
			if err = helm.Pull(c.Source, c.Path, c.Version, tmpHelmDir); err != nil {
				return err
			}

			// Create the component directory -- deleting if it already exists
			if err != nil {
				return err
			}
			if err := os.RemoveAll(helmRepoPath); err != nil {
				return err
			}

			// ensure the parent directory exists
			if err := os.MkdirAll(filepath.Dir(helmRepoPath), 0755); err != nil {
				return err
			}

			// Move the extracted chart from tmp to the helm_repos
			extractedChartPath := path.Join(tmpHelmDir, c.Path)
			if err := os.Rename(extractedChartPath, helmRepoPath); err != nil {
				return err
			}
		case "git":
			// Clone whole repo into helm repo path
			logger.Info(emoji.Sprintf(":helicopter: Component '%s' requesting helm chart in path '%s' from git repository '%s'", c.Name, c.Source, c.PhysicalPath))
			cloneOpts := &git.CloneOpts{
				URL:    c.Source,
				SHA:    c.Version,
				Branch: c.Branch,
				Into:   helmRepoPath,
			}
			if err = git.Clone(cloneOpts); err != nil {
				return err
			}
			// Update chart dependencies in chart path -- this is manually done here but automatically done in downloadChart in the case of `method: helm`
			chartPath, err := hg.getChartPath(c)
			if err != nil {
				return err
			}
			if err = helm.DependencyUpdate(chartPath); err != nil {
				return err
			}
		}
	}

	return err
}
