package models

import "encoding/json"

type Component struct {
	Name          string
	Type          string
	Subcomponents []Subcomponent
}

func ParseComponentFromJson(componentJson []byte) (component *Component) {
	if err := json.Unmarshal(componentJson, &component); err != nil {
		return nil
	}

	return component
}
