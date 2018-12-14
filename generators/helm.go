package generators

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"path"
	"path/filepath"

	"github.com/Microsoft/marina/core"
	"gopkg.in/yaml.v2"
)

func GenerateHelmComponent(component *core.Component) (definition string, err error) {
	fmt.Printf("ooo generating component %s with helm with repo %s\n", component.Name, component.Repo)

	configYaml, err := yaml.Marshal(&component.Config.Config)
	if err != nil {
		return "", err
	}

	helmRepoPath := path.Join(component.PhysicalPath, "repo")
	absHelmRepoPath, err := filepath.Abs(helmRepoPath)
	chartPath := path.Join(absHelmRepoPath, component.Path)
	absCustomValuesPath := path.Join(chartPath, "overriddenValues.yaml")

	ioutil.WriteFile(absCustomValuesPath, configYaml, 0644)

	volumeMount := fmt.Sprintf("%s:/app/chart", chartPath)

	// docker run --rm -v `pwd`/fluentd-elasticsearch:/app/chart alpine/helm:latest template /app/chart --set 'namespace=prom,master.persistence.size=4Gi'
	output, err := exec.Command("docker", "run", "--rm", "-v", volumeMount, "alpine/helm:latest", "template", "/app/chart", "--values", "/app/chart/overriddenValues.yaml").Output()

	if err != nil {
		return "", err
	}

	return string(output), nil
}

func InstallHelmComponent(component *core.Component) (err error) {
	helmRepoPath := path.Join(component.PhysicalPath, "repo")
	if err := exec.Command("rm", "-rf", helmRepoPath).Run(); err != nil {
		return err
	}

	fmt.Printf("vvv install helm repo %s for %s into %s\n", component.Repo, component.Name, helmRepoPath)
	return exec.Command("git", "clone", component.Repo, helmRepoPath, "--depth", "1").Run()
}
