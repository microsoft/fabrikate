package generatable

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/microsoft/fabrikate/internal/helm"
	"github.com/timfpark/yaml"
)

type Helm struct {
	ComponentPath []string // list representation of the component structure; used to generate the generation path
	ChartPath     string   // path to directory containing Chart.yaml. Not to actual chart.yaml file
	Values        map[string]interface{}
}

func (h Helm) Validate() error {
	return nil
}

func (h Helm) Generate() error {
	// write values.yaml to a temporary file
	valuesFile, err := ioutil.TempFile("", "fabrikate")
	if err != nil {
		return fmt.Errorf(`error creating temporary helm values files: %w`, err)
	}
	defer os.Remove(valuesFile.Name())
	valueBytes, err := yaml.Marshal(h.Values)
	if err != nil {
		return fmt.Errorf(`error marshalling helm values to temporary file: %w`, err)
	}
	if _, err := valuesFile.Write(valueBytes); err != nil {
		return fmt.Errorf(`error writing temporary helm values file: %w`, err)
	}
	if err := valuesFile.Close(); err != nil {
		return fmt.Errorf(`error writing temporary helm values file: %w`, err)
	}

	// run `helm template`
	template, err := helm.Template(helm.TemplateOptions{
		Chart:  h.ChartPath,
		Values: []string{valuesFile.Name()},
	})
	if err != nil {
		return fmt.Errorf(`helm template error: %w`, err)
	}

	// remove existing generation
	generatePath := h.GetGeneratePath()
	if err := os.Remove(generatePath); err != nil {
		return err
	}

	// write out template
	if err := ioutil.WriteFile(generatePath, []byte(template), 0755); err != nil {
		return err
	}

	return nil
}

func (h Helm) GetGeneratePath() string {
	componentName := strings.Join(h.ComponentPath, "__")
	pathSegments := []string{generateDirName, componentName}

	return path.Join(pathSegments...) + ".yaml"
}
