package util

import (
	"os"
	"path/filepath"
	"regexp"
)

// ListComponentInstallDirectories returns all subdirectories in `directory` which have have the name
// "components" or "helm_repos"; this is mainly used as a helper function for cleaning up test `Install`s
func ListComponentInstallDirectories(directory string) (componentDirs []string, err error) {
	err = filepath.Walk(directory, func(path string, file os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if file.IsDir() {
			if match, err := regexp.MatchString("/(components|helm_repos)$", path); match && err == nil {
				componentDirs = append(componentDirs, path)
			}
		}
		return nil
	})

	return componentDirs, err
}

// UninstallComponents uninstalls any components in any subdirectory under `path`.
// Equivalent to `rm -rf **/components **/helm_repos`
func UninstallComponents(path string) (err error) {
	dirsToClean, err := ListComponentInstallDirectories(path)
	if err != nil {
		return err
	}
	for _, dir := range dirsToClean {
		if err = os.RemoveAll(dir); err != nil {
			return err
		}
	}
	return err
}
