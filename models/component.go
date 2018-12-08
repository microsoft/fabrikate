package models

import "encoding/json"

type Component struct {
	Name          string
	Type          string
	Subcomponents []Subcomponent
}

func ParseComponentFromJson(componentJson []byte) (component Component) {
	json.Unmarshal(componentJson, &component)

	return component
}
