package core

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path"

	"github.com/timfpark/conjungo"
	yaml "github.com/timfpark/yaml"
)

type ComponentConfig struct {
	Path          string                     `yaml:"-" json:"-"`
	Serialization string                     `yaml:"-" json:"-"`
	Config        map[string]interface{}     `yaml:"config,omitempty" json:"config,omitempty"`
	Subcomponents map[string]ComponentConfig `yaml:"subcomponents,omitempty" json:"subcomponents,omitempty"`
}

func NewComponentConfig() ComponentConfig {
	return ComponentConfig{
		Config:        map[string]interface{}{},
		Subcomponents: map[string]ComponentConfig{},
	}
}

func (cc *ComponentConfig) GetPath(environment string) string {
	configFilename := fmt.Sprintf("config/%s.%s", environment, cc.Serialization)
	return path.Join(cc.Path, configFilename)
}

func (cc *ComponentConfig) UnmarshalJSONConfig(environment string) (err error) {
	cc.Serialization = "json"
	return UnmarshalFile(cc.GetPath(environment), json.Unmarshal, &cc)
}

func (cc *ComponentConfig) UnmarshalYAMLConfig(environment string) (err error) {
	cc.Serialization = "yaml"
	return UnmarshalFile(cc.GetPath(environment), yaml.Unmarshal, &cc)
}

func (cc *ComponentConfig) MergeConfigFile(environment string) (err error) {
	componentConfig := NewComponentConfig()

	if err := componentConfig.Load(environment); err != nil {
		return err
	}

	return cc.Merge(componentConfig)
}

func (cc *ComponentConfig) Load(environment string) (err error) {
	err = cc.UnmarshalYAMLConfig(environment)

	// fall back to looking for JSON if loading YAML fails.
	if err != nil {
		err = cc.UnmarshalJSONConfig(environment)

		if err != nil {
			// couldn't find any config files, so default back to yaml serialization
			cc.Serialization = "yaml"
		}
	}

	return nil
}

func (cc *ComponentConfig) SetComponentConfig(path []string, value string) {
	configLevel := cc.Config

	for levelIndex, pathPart := range path {
		// if this key is not the final one, we need to decend in the config.
		if levelIndex < len(path)-1 {
			if _, ok := configLevel[pathPart]; !ok {
				configLevel[pathPart] = map[string]interface{}{}
			}

			configLevel = configLevel[pathPart].(map[string]interface{})
		} else {
			configLevel[pathPart] = value
		}
	}
}

func (cc *ComponentConfig) SetConfig(subcomponentPath []string, path []string, value string) {
	subcomponentConfig := *cc
	for _, subcomponentName := range subcomponentPath {
		if subcomponentConfig.Subcomponents == nil {
			subcomponentConfig.Subcomponents = map[string]ComponentConfig{}
		}

		if _, ok := subcomponentConfig.Subcomponents[subcomponentName]; !ok {
			subcomponentConfig.Subcomponents[subcomponentName] = NewComponentConfig()
		}

		subcomponentConfig = subcomponentConfig.Subcomponents[subcomponentName]
	}

	subcomponentConfig.SetComponentConfig(path, value)
}

func (cc *ComponentConfig) Merge(newConfig ComponentConfig) (err error) {
	options := conjungo.NewOptions()
	options.Overwrite = false

	err = conjungo.Merge(cc, newConfig, options)

	return err
}

func (cc *ComponentConfig) Write(environment string) (err error) {
	var marshaledConfig []byte

	if cc.Serialization == "json" {
		marshaledConfig, err = json.MarshalIndent(cc, "", "  ")
	} else {
		marshaledConfig, err = yaml.Marshal(cc)
	}

	return ioutil.WriteFile(cc.GetPath(environment), marshaledConfig, 0644)
}
