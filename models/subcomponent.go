package models

import "fmt"

type Subcomponent struct {
	Name   string
	Method string
	Source string
}

func (sc *Subcomponent) RelativePathTo() string {
	if sc.Method == "git" {
		return fmt.Sprintf("components/%s", sc.Name)
	} else {
		return sc.Name
	}
}
