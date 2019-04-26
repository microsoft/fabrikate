package core

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/kyokomi/emoji"
	log "github.com/sirupsen/logrus"
	"github.com/timfpark/conjungo"
	yaml "github.com/timfpark/yaml"
)

// ComponentConfig documentation: https://github.com/Microsoft/fabrikate/blob/master/docs/config.md
type ComponentConfig struct {
	Path            string                     `yaml:"-" json:"-"`
	Serialization   string                     `yaml:"-" json:"-"`
	Namespace       string                     `yaml:"namespace,omitempty" json:"namespace,omitempty"`
	InjectNamespace bool                       `yaml:"injectNamespace,omitempty" json:"injectNamespace,omitempty"`
	Config          map[string]interface{}     `yaml:"config,omitempty" json:"config,omitempty"`
	Subcomponents   map[string]ComponentConfig `yaml:"subcomponents,omitempty" json:"subcomponents,omitempty"`
}

// NewComponentConfig creates a ComponentConfig at the passed path.
func NewComponentConfig(path string) ComponentConfig {
	return ComponentConfig{
		Path:          path,
		Config:        map[string]interface{}{},
		Subcomponents: map[string]ComponentConfig{},
	}
}

// GetPath returns the path to the config file for the specified environment.
func (cc *ComponentConfig) GetPath(environment string) string {
	configFilename := fmt.Sprintf("config/%s.%s", environment, cc.Serialization)
	return path.Join(cc.Path, configFilename)
}

// UnmarshalJSONConfig unmarshals the JSON config file for the specified environment.
func (cc *ComponentConfig) UnmarshalJSONConfig(environment string) (err error) {
	cc.Serialization = "json"
	return UnmarshalFile(cc.GetPath(environment), json.Unmarshal, &cc)
}

// UnmarshalYAMLConfig unmarshals the YAML config file for the specified environment.
func (cc *ComponentConfig) UnmarshalYAMLConfig(environment string) (err error) {
	cc.Serialization = "yaml"
	return UnmarshalFile(cc.GetPath(environment), yaml.Unmarshal, &cc)
}

// MergeConfigFile loads the config for the specified environment and path and
// merges it with the current set of config.
func (cc *ComponentConfig) MergeConfigFile(path string, environment string) (err error) {
	componentConfig := NewComponentConfig(path)
	if err := componentConfig.Load(environment); err != nil {
		return err
	}

	return cc.Merge(componentConfig)
}

// Load loads the config for the specified environment.
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

// HasComponentConfig checks if the component contains the given component configuration.
// The given component is specified via a configuration `path`.
// Returns true if it contains it, otherwise it returns false.
func (cc *ComponentConfig) HasComponentConfig(path []string) bool {
	configLevel := cc.Config

	for levelIndex, pathPart := range path {
		// if this key is not the final one, we need to decend in the config.
		if _, ok := configLevel[pathPart]; !ok {
			return false
		}

		if levelIndex < len(path)-1 {
			configLevel = configLevel[pathPart].(map[string]interface{})
		}
	}

	return true
}

// SetComponentConfig sets the `value` of the given configuration setting.
// The configuration setting is indicated via a configuration `path`.
func (cc *ComponentConfig) SetComponentConfig(path []string, value string) {
	configLevel := cc.Config
	createdNewConfig := false

	for levelIndex, pathPart := range path {
		// if this key is not the final one, we need to decend in the config.
		if levelIndex < len(path)-1 {
			if _, ok := configLevel[pathPart]; !ok {
				createdNewConfig = true
				configLevel[pathPart] = map[string]interface{}{}
			}

			configLevel = configLevel[pathPart].(map[string]interface{})
		} else {
			if createdNewConfig {
				log.Info(emoji.Sprintf(":seedling: Created new value for %s", strings.Join(path, ".")))
			}
			configLevel[pathPart] = value
		}
	}
}

// GetSubcomponentConfig returns the subcomponent config of the given component.
// If the subcomponent does not exist, it creates it
//
// Returns the subcomponent config
func (cc *ComponentConfig) GetSubcomponentConfig(subcomponentPath []string) (subcomponentConfig ComponentConfig) {
	subcomponentConfig = *cc
	for _, subcomponentName := range subcomponentPath {
		if subcomponentConfig.Subcomponents == nil {
			subcomponentConfig.Subcomponents = map[string]ComponentConfig{}
		}

		if _, ok := subcomponentConfig.Subcomponents[subcomponentName]; !ok {
			log.Info(emoji.Sprintf(":seedling: Creating new subcomponent configuration for %s", subcomponentName))
			subcomponentConfig.Subcomponents[subcomponentName] = NewComponentConfig(".")
		}

		subcomponentConfig = subcomponentConfig.Subcomponents[subcomponentName]
	}

	return subcomponentConfig
}

// HasSubcomponentConfig checks if a component contains the given subcomponents of the `subcomponentPath`
//
// Returns true if it contains the subcomponent, otherwise it returns false
func (cc *ComponentConfig) HasSubcomponentConfig(subcomponentPath []string) bool {
	subcomponentConfig := *cc

	for _, subcomponentName := range subcomponentPath {
		if subcomponentConfig.Subcomponents == nil {
			return false
		}

		if _, ok := subcomponentConfig.Subcomponents[subcomponentName]; !ok {
			return false
		}

		subcomponentConfig = subcomponentConfig.Subcomponents[subcomponentName]
	}

	return true
}

// SetConfig sets or creates the configuration `value` for the given `subcomponentPath`.
func (cc *ComponentConfig) SetConfig(subcomponentPath []string, path []string, value string) {
	subcomponentConfig := cc.GetSubcomponentConfig(subcomponentPath)
	subcomponentConfig.SetComponentConfig(path, value)
}

// MergeNamespaces merges the namespaces between the componentConfig passed and this
// ComponentConfig.
func (cc *ComponentConfig) MergeNamespaces(newConfig ComponentConfig) ComponentConfig {
	if cc.Namespace == "" {
		cc.Namespace = newConfig.Namespace
		cc.InjectNamespace = newConfig.InjectNamespace
	}

	for key, config := range cc.Subcomponents {
		cc.Subcomponents[key] = config.MergeNamespaces(newConfig.Subcomponents[key])
	}

	return *cc
}

// Merge merges the config (and the namespace spec) between the passed componentConfig
// and this componentConfig.  In the case of conflicts, this componentConfig wins.
func (cc *ComponentConfig) Merge(newConfig ComponentConfig) (err error) {
	options := conjungo.NewOptions()
	options.Overwrite = false

	err = conjungo.Merge(cc, newConfig, options)

	cc.MergeNamespaces(newConfig)

	return err
}

// Write writes this componentConfig to a file using the serialization specified in
// cc.Serialization.
func (cc *ComponentConfig) Write(environment string) (err error) {
	var marshaledConfig []byte

	_ = os.Mkdir(cc.Path, os.ModePerm)
	_ = os.Mkdir(path.Join(cc.Path, "config"), os.ModePerm)

	if cc.Serialization == "json" {
		marshaledConfig, err = json.MarshalIndent(cc, "", "  ")
	} else {
		marshaledConfig, err = yaml.Marshal(cc)
	}

	if err != nil {
		return err
	}

	return ioutil.WriteFile(cc.GetPath(environment), marshaledConfig, 0644)
}
