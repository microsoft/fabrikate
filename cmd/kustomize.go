package cmd

import (
	"fmt"
	"io/ioutil"
	"path"

	"github.com/kyokomi/emoji"
	"github.com/microsoft/fabrikate/core"
	"github.com/microsoft/fabrikate/logger"
	"github.com/timfpark/yaml"
)

type kustomization struct {
	Kind       string   `yaml:"kind,omitempty"`
	APIVersion string   `yaml:"apiversion,omitempty"`
	Resources  []string `yaml:"resources,omitempty"`
}

const kustomizationFileName = "kustomization.yaml"
const defaultKind = "Kustomization"
const defaultAPIVersion = "kustomize.config.k8s.io/v1beta1"

func (k *kustomization) setDefaultEmptyKustomization() {
	if k.Kind == "" {
		k.Kind = defaultKind
	}
	if k.APIVersion == "" {
		k.APIVersion = defaultAPIVersion
	}
	if k.Resources == nil {
		k.Resources = make([]string, 0)
	}
}

func (k *kustomization) addKustomizationResource(component core.Component) {
	componentYAMLFilename := fmt.Sprintf("%s.yaml", component.Name)
	componentYAMLFilepath := path.Join(component.LogicalPath, componentYAMLFilename)
	logger.Debug(emoji.Sprintf(":truck: Adding resource %s to %s", componentYAMLFilepath, kustomizationFileName))
	k.Resources = append(k.Resources, componentYAMLFilepath)
}

func writeKustomizationFile(generationPath string, kustomizationBytes []byte) (err error) {
	kustomizationFile := path.Join(generationPath, kustomizationFileName)
	logger.Info(emoji.Sprintf(":floppy_disk: Writing %s", kustomizationFile))
	return ioutil.WriteFile(kustomizationFile, kustomizationBytes, 0644)
}

func createKustomizationFile(generationPath string, components []core.Component) (err error) {
	kustomization := kustomization{}
	kustomization.setDefaultEmptyKustomization()

	for _, component := range components {
		kustomization.addKustomizationResource(component)
	}

	kustomizationBytes, err := yaml.Marshal(kustomization)

	if err != nil {
		return err
	}

	err = writeKustomizationFile(generationPath, kustomizationBytes)

	if err != nil {
		return err
	}

	return nil
}
