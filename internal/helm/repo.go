package helm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
)

// RepoListEntry is a single entry from the output of
// `helm repo list --output json`
type RepoListEntry struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

// RepoList lists all repositories currently in the host Helm client
func RepoList() (list []RepoListEntry, err error) {
	lock.RLock()
	defer lock.RUnlock()

	listCmd := exec.Command("helm", "repo", "list", "--output", "json")
	var stdout, stderr bytes.Buffer
	listCmd.Stdout = &stdout
	listCmd.Stderr = &stderr
	if err := listCmd.Run(); err != nil {
		return list, fmt.Errorf("%v: %v", err, stderr.String())
	}

	if err := json.Unmarshal(stdout.Bytes(), &list); err != nil {
		return list, err
	}

	return list, err
}

// RepoAdd adds a helm repository of `name` pointing to `url` to the host Helm
// client
func RepoAdd(name string, url string) error {
	lock.Lock()
	defer lock.Unlock()

	addCmd := exec.Command("helm", "repo", "add", name, url)
	var stdout, stderr bytes.Buffer
	addCmd.Stdout = &stdout
	addCmd.Stderr = &stderr
	if err := addCmd.Run(); err != nil {
		return fmt.Errorf("%v: %v", err, stderr.String())
	}

	return nil
}

// RepoRemove attempts to remove the helm repository of `name` from the host
// helm client
func RepoRemove(name string) error {
	lock.Lock()
	defer lock.Unlock()

	removeCmd := exec.Command("helm", "repo", "remove", name)
	var stdout, stderr bytes.Buffer
	removeCmd.Stdout = &stdout
	removeCmd.Stderr = &stderr
	if err := removeCmd.Run(); err != nil {
		return fmt.Errorf("%v: %v", err, stderr.String())
	}

	return nil
}

// FindRepoNameByURL attempts to search for an existing helm repository on the
// the host matching the provided URL.
// Will return the the name of the repo if found or empty string if not.
// Errors when unable to parse the host repository list.
func FindRepoNameByURL(URL string) (string, error) {
	repositories, err := RepoList()
	if err != nil {
		return "", err
	}
	for _, entry := range repositories {
		if entry.URL == URL {
			return entry.Name, nil
		}
	}

	return "", nil
}
