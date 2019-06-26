package core

import (
	"fmt"
	"os/exec"
	"regexp"

	"github.com/kyokomi/emoji"
	log "github.com/sirupsen/logrus"
)

// CloneRepo is a helper func to centralize cloning a repository with the spec provided by its arguments.
func CloneRepo(repo string, commit string, intoPath string, branch string, accessToken string) (err error) {
	if accessToken != "" {
		// Only match when the repo string does not contain a an access token already
		// "(https?)://(?!(.+:)?.+@)(.+)" would be preferred but go does not support negative lookahead
		pattern, err := regexp.Compile("^(https?)://([^@]+@)?(.+)$")
		if err != nil {
			return err
		}
		// If match is found, inject the access token into the repo string
		matches := pattern.FindStringSubmatch(repo)
		if matches != nil {
			protocol := matches[1]
			// credentialsWithAtSign := matches[2]
			cleanedRepoString := matches[3]
			repo = fmt.Sprintf("%v://%v@%v", protocol, accessToken, cleanedRepoString)
		}
	}

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
	cloneCommand := exec.Command("git", cloneArgs...)
	cloneCommand.Env = append(cloneCommand.Env, "GIT_TERMINAL_PROMPT=0") // tell git to fail if it asks for credentials
	if err = cloneCommand.Run(); err != nil {
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
