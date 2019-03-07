package core

import (
	"os/exec"

	"github.com/kyokomi/emoji"
	log "github.com/sirupsen/logrus"
)

func CloneRepo(repo string, commit string, intoPath string, branch string) (err error) {
	cloneArgs := []string{
		"clone",
		repo,
	}

	if len(commit) == 0 {
		log.Println(emoji.Sprintf(":helicopter: component requested latest commit: fast cloning at --depth 1"))

		cloneArgs = append(cloneArgs, "--depth", "1")
	} else {
		log.Println(emoji.Sprintf(":helicopter: component requested specific commit: need full clone"))
	}

	if len(branch) != 0 {
		cloneArgs = append(cloneArgs, "--branch", branch)
	}

	cloneArgs = append(cloneArgs, intoPath)

	if err = exec.Command("git", cloneArgs...).Run(); err != nil {
		return err
	}

	if len(commit) != 0 {
		log.Println(emoji.Sprintf(":helicopter: performing checkout at commit %s", commit))
		checkoutCommit := exec.Command("git", "checkout", commit)
		checkoutCommit.Dir = intoPath
		if err = checkoutCommit.Run(); err != nil {
			return err
		}
	}

	return nil
}
