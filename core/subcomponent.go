package core

import (
	"fmt"
	"os/exec"
	"path"
)

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

func (sc *Subcomponent) Install(componentPath string) (err error) {
	if sc.Method == "git" {
		componentsPath := fmt.Sprintf("%s/components", componentPath)
		if err := exec.Command("mkdir", "-p", componentsPath).Run(); err != nil {
			return err
		}

		subcomponentPath := path.Join(componentPath, sc.RelativePathTo())
		if err = exec.Command("rm", "-rf", subcomponentPath).Run(); err != nil {
			return err
		}

		fmt.Printf("vvv installing component %s with git from %s\n", sc.Name, sc.Source)
		if err = exec.Command("git", "clone", sc.Source, subcomponentPath).Run(); err != nil {
			return err
		}
	}

	return nil
}
