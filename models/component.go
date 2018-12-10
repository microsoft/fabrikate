package models

import "encoding/json"

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
