package core

import "github.com/InVisionApp/conjungo"

type ComponentConfig struct {
	Config        map[string]interface{}
	Subcomponents map[string]ComponentConfig
}

func (cc *ComponentConfig) Merge(newConfig ComponentConfig) (err error) {
	options := conjungo.NewOptions()
	options.Overwrite = false

	return conjungo.Merge(cc, newConfig, options)
}
