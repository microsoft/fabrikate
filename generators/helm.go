package generators

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/kyokomi/emoji"
	"github.com/microsoft/fabrikate/core"
	log "github.com/sirupsen/logrus"
	"github.com/timfpark/yaml"
)

// HelmGenerator provides 'helm generate' generator functionality to Fabrikate
type HelmGenerator struct{}

func addNamespaceToManifests(manifests string, namespace string) (namespacedManifests string, err error) {
	splitManifest := strings.Split(manifests, "\n---")

	for _, manifest := range splitManifest {
		parsedManifest := make(map[interface{}]interface{})
		err := yaml.Unmarshal([]byte(manifest), &parsedManifest)
		if err != nil {
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

func (hg *HelmGenerator) makeHelmRepoPath(component *core.Component) string {
	if component.Method != "git" {
		return component.PhysicalPath
	}

	return path.Join(component.PhysicalPath, "helm_repos", component.Name)
}

// Generate returns the helm templated manifests specified by this component.
func (hg *HelmGenerator) Generate(component *core.Component) (manifest string, err error) {
	log.Println(emoji.Sprintf(":truck: Generating component '%s' with helm with repo %s", component.Name, component.Source))

	configYaml, err := yaml.Marshal(&component.Config.Config)
	if err != nil {
		log.Errorf("Marshalling config yaml for helm generated component '%s' failed with: %s\n", component.Name, err.Error())
		return "", err
	}

	helmRepoPath := hg.makeHelmRepoPath(component)
	absHelmRepoPath, err := filepath.Abs(helmRepoPath)
	if err != nil {
		return "", err
	}

	chartPath := path.Join(absHelmRepoPath, component.Path)
	uuid, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}

	overriddenValuesFileName := fmt.Sprintf("overriddenValues-%s.yaml", uuid.String())
	absOverriddenPath := path.Join(chartPath, overriddenValuesFileName)

	log.Debugf("Writing config %s to %s\n", configYaml, absOverriddenPath)
	err = ioutil.WriteFile(absOverriddenPath, configYaml, 0644)
	if err != nil {
		return "", err
	}

	name := component.Name

	namespace := "default"
	if component.Config.Namespace != "" {
		namespace = component.Config.Namespace
	}

	output, err := exec.Command("helm", "template", chartPath, "--values", absOverriddenPath, "--name", name, "--namespace", namespace).Output()

	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			log.Errorf("helm template failed with: %s\n", ee.Stderr)
			_ = exec.Command("rm", absOverriddenPath).Run()
			return "", err
		}
	}

	stringManifests := string(output)

	// helm template does not inject namespace unless chart directly provides support for it: https://github.com/helm/helm/issues/3553
	// some helm templates expect Tiller to inject namespace, so enable Fabrikate component designer to
	// opt into injecting these namespaces manually.  We should reassess if this is necessary after Helm 3 is released and client side
	// templating really becomes a first class function in Helm.
	if component.Config.InjectNamespace && component.Config.Namespace != "" {
		stringManifests, err = addNamespaceToManifests(stringManifests, component.Config.Namespace)
	}

	_ = exec.Command("rm", absOverriddenPath).Run()

	return stringManifests, err
}

// Install installs the helm chart specified by the passed component and performs any
// helm lifecycle events needed.
func (hg *HelmGenerator) Install(component *core.Component, accessTokens map[string]string) (err error) {
	if len(component.Source) == 0 || component.Method != "git" {
		return nil
	}

	helmRepoPath := hg.makeHelmRepoPath(component)
	if err := exec.Command("rm", "-rf", helmRepoPath).Run(); err != nil {
		return err
	}

	if err := exec.Command("mkdir", "-p", helmRepoPath).Run(); err != nil {
		return err
	}

	log.Println(emoji.Sprintf(":helicopter: Installing helm repo %s for %s into %s", component.Source, component.Name, helmRepoPath))

	// Access token lookup
	accessToken := ""
	if foundToken, ok := accessTokens[component.Source]; ok {
		accessToken = foundToken
	}

	if err = core.CloneRepo(component.Source, component.Version, helmRepoPath, component.Branch, accessToken); err != nil {
		return err
	}

	absHelmRepoPath, err := filepath.Abs(helmRepoPath)
	if err != nil {
		return err
	}

	chartPath := path.Join(absHelmRepoPath, component.Path)

	for name, url := range component.Repositories {
		log.Println(emoji.Sprintf(":helicopter: Adding helm repo '%s' at %s for component '%s'", name, url, component.Name))
		if err = exec.Command("helm", "repo", "add", name, url).Run(); err != nil {
			return err
		}
	}

	log.Println(emoji.Sprintf(":helicopter: Updating helm chart's dependencies for component '%s'", component.Name))
	err = exec.Command("helm", "dependency", "update", chartPath).Run()

	if err != nil {
		log.Errorf("Updating chart dependencies failed\n")
		log.Errorf("Run 'helm dependency update %s' for more error details.\n", chartPath)
	}

	return err
}
