package helm

import (
	"bytes"
	"fmt"
	"os/exec"
)

// Pull will do a `helm pull` for the target chart and extract the chart to
// `into`.
func Pull(repoURL string, chart string, version string, into string) error {
	// check if existing repo with same URL in host client
	existingRepo, _ := FindRepoNameByURL(repoURL)
	if len(existingRepo) > 0 {
		chart = existingRepo + "/" + chart
	}

	// arguments don't include --repo by default
	pullArgs := []string{"pull", chart,
		"--version", version,
		"--untar",
		"--untardir", into}

	// use the --repo option to pull directly from URL if repo not on host Helm
	if len(existingRepo) == 0 {
		pullArgs = append(pullArgs, "--repo", repoURL)
	}

	cmd := exec.Command("helm", pullArgs...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%v: %v", err, stderr.String())
	}

	return nil
}
