package installable

import (
	"os"
	"path"
)

type Local struct {
	Root string
}

func (l Local) Install() error {
	componentPath := path.Join(l.Root)
	if _, err := os.Stat(componentPath); os.IsNotExist(err) {
		return err
	}

	return nil
}

func (l Local) GetInstallPath() (string, error) {
	return l.Root, nil
}

func (l Local) Validate() error {
	return nil
}
