package generators

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/Microsoft/fabrikate/core"
	"github.com/kyokomi/emoji"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type HelmGenerator struct{}

func AddNamespaceToManifests(manifests string, namespace string) (namespacedManifests string, err error) {
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
			metadataMap["namespace"] = namespace
		}

		updatedManifest, err := yaml.Marshal(&parsedManifest)
		if err != nil {
			return "", err
		}

		namespacedManifests += fmt.Sprintf("---\n%s\n", updatedManifest)
	}

	return namespacedManifests, nil
}

func (hg *HelmGenerator) MakeHelmRepoPath(component *core.Component) string {
	if len(component.Repo) == 0 {
		return component.PhysicalPath
	} else {
		return path.Join(component.PhysicalPath, "helm_repos", component.Name)
	}
}

func (hg *HelmGenerator) Generate(component *core.Component) (manifest string, err error) {
	log.Println(emoji.Sprintf(":truck: generating component '%s' with helm with repo %s", component.Name, component.Repo))

	configYaml, err := yaml.Marshal(&component.Config.Config)
	if err != nil {
		log.Errorf("marshalling config yaml for helm generated component '%s' failed with: %s\n", component.Name, err.Error())
		return "", err
	}

	helmRepoPath := hg.MakeHelmRepoPath(component)
	absHelmRepoPath, err := filepath.Abs(helmRepoPath)
	if err != nil {
		return "", err
	}

	chartPath := path.Join(absHelmRepoPath, component.Path)
	absOverriddenPath := path.Join(chartPath, "overriddenValues.yaml")

	log.Debugf("writing config %s to %s\n", configYaml, absOverriddenPath)
	err = ioutil.WriteFile(absOverriddenPath, configYaml, 0644)
	if err != nil {
		return "", err
	}

	name := component.Name
	if component.Config.Config["name"] != nil {
		name = component.Config.Config["name"].(string)
	}

	namespace := "default"
	if component.Config.Config["namespace"] != nil {
		namespace = component.Config.Config["namespace"].(string)
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

	// some helm templates expect install to inject namespace, so if namespace doesn't exist on resource manifests, manually inject it.
	if component.Config.Config["namespace"] != nil {
		stringManifests, err = AddNamespaceToManifests(stringManifests, component.Config.Config["namespace"].(string))
	}

	_ = exec.Command("rm", absOverriddenPath).Run()

	return stringManifests, err
}

func (hg *HelmGenerator) Install(component *core.Component) (err error) {
	if len(component.Repo) == 0 {
		return nil
	}

	helmRepoPath := hg.MakeHelmRepoPath(component)
	if err := exec.Command("rm", "-rf", helmRepoPath).Run(); err != nil {
		return err
	}

	if err := exec.Command("mkdir", "-p", helmRepoPath).Run(); err != nil {
		return err
	}

	gitArguments, versionString := component.BuildGitCloneArguments(helmRepoPath)

	gitArguments = append(gitArguments, "--depth", "1")

	log.Println(emoji.Sprintf(":helicopter: install helm repo %s for %s into %s%s", component.Repo, component.Name, helmRepoPath, versionString))
	if err := exec.Command("git", gitArguments...).Run(); err != nil {
		return err
	}

	absHelmRepoPath, err := filepath.Abs(helmRepoPath)
	if err != nil {
		return err
	}

	chartPath := path.Join(absHelmRepoPath, component.Path)

	log.Println(emoji.Sprintf(":helicopter: updating helm chart's dependencies for %s", component.Name))

	err = exec.Command("helm", "dependency", "update", chartPath).Run()

	if err != nil {
		log.Errorf("updating chart dependencies failed\n")
		log.Errorf("run 'helm dependency update %s' for more error details.\n", chartPath)
	}

	return err
}
