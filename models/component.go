package models

import (
	"encoding/json"
)

type Component struct {
	Name          string
	Type          string
	Subcomponents []Subcomponent
}

func ParseComponentFromJson(componentJson []byte) (component *Component, err error) {
	if err := json.Unmarshal(componentJson, &component); err != nil {
		return nil, err
	}

	return component, nil
}

func (c *Component) Install(path string) (err error) {
	for _, subcomponent := range c.Subcomponents {
		if err := subcomponent.Install(path); err != nil {
			return err
		}
	}

	return err
}
