package core

import (
	"github.com/imdario/mergo"
)

type ComponentConfig struct {
	Config        map[string]interface{}
	Subcomponents map[string]ComponentConfig
}

func (cc *ComponentConfig) Merge(from ComponentConfig) (err error) {
	return mergo.Merge(cc, from)
}
