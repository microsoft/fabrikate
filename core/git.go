package core

import (
	"os/exec"

	"github.com/kyokomi/emoji"
	log "github.com/sirupsen/logrus"
)

// CloneRepo is a helper func to centralize cloning a repo with the spec provided by its arguments.
func CloneRepo(repo string, commit string, intoPath string, branch string) (err error) {
	cloneArgs := []string{
		"clone",
		repo,
	}

	if len(commit) == 0 {
		log.Println(emoji.Sprintf(":helicopter: Component requested latest commit: fast cloning at --depth 1"))

		cloneArgs = append(cloneArgs, "--depth", "1")
	} else {
		log.Println(emoji.Sprintf(":helicopter: Component requested commit '%s': need full clone", commit))
	}

	if len(branch) != 0 {
		log.Println(emoji.Sprintf(":helicopter: Component requested branch '%s'", branch))
		cloneArgs = append(cloneArgs, "--branch", branch)
	}

	cloneArgs = append(cloneArgs, intoPath)

	if err = exec.Command("git", cloneArgs...).Run(); err != nil {
		return err
	}

	if len(commit) != 0 {
		log.Println(emoji.Sprintf(":helicopter: Performing checkout at commit %s", commit))
		checkoutCommit := exec.Command("git", "checkout", commit)
		checkoutCommit.Dir = intoPath
		if err = checkoutCommit.Run(); err != nil {
			return err
		}
	}

	return nil
}
