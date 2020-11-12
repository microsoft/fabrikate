package installable

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path"
	"sync"

	"github.com/microsoft/fabrikate/internal/url"
)

type Git struct {
	URL    string
	SHA    string
	Branch string
}

func (g Git) Install() error {
	// deleting if it already exists
	componentPath, err := g.GetInstallPath()
	if err != nil {
		return err
	}
	if err := os.RemoveAll(componentPath); err != nil {
		return err
	}
	// clone the repo
	if err := clone(g.URL, componentPath); err != nil {
		return err
	}
	// checkout target Branch
	if g.Branch != "" {
		if err := checkout(componentPath, g.Branch); err != nil {
			return err
		}
	}

	// checkout target SHA
	if g.SHA != "" {
		if err := checkout(componentPath, g.SHA); err != nil {
			return err
		}
	}

	return nil
}

func (g Git) GetInstallPath() (string, error) {
	urlPath, err := url.ToPath(g.URL)
	if err != nil {
		return "", err
	}

	var version string
	if g.SHA != "" {
		version = g.SHA
	} else if g.Branch != "" {
		version = g.Branch
	} else {
		version = "latest"
	}

	componentPath := path.Join(installDirName, urlPath, version)
	return componentPath, nil
}

func (g Git) Validate() error {
	if g.URL == "" {
		return fmt.Errorf(`URL must be non-zero length`)
	}
	if g.SHA != "" && g.Branch != "" {
		return fmt.Errorf(`Only one of SHA or Branch can be provided, "%v" and "%v" provided respectively`, g.SHA, g.Branch)
	}

	return nil
}

//------------------------------------------------------------------------------
// Git Helpers
//------------------------------------------------------------------------------

type coordinator struct {
	coordinator *sync.Mutex              // lock to ensure only one write has access to locks at a time
	nodes       map[string]*sync.RWMutex // each lock determines if the key has been successfully cloned
}

var pathCoordinator = coordinator{
	coordinator: &sync.Mutex{},
	nodes:       map[string]*sync.RWMutex{},
}

// clone performs a `git clone <repo> <dir>`
func clone(repo string, dir string) error {
	coordinator := pathCoordinator.coordinator
	nodes := pathCoordinator.nodes
	coordinator.Lock() // establish a read lock so we can safely read from the map of locks

	// If one exists, we just need to wait for it to become free; establish a lock
	// and immediately release
	if node, exists := nodes[dir]; exists {
		node.RLock()
		coordinator.Unlock()
		node.RUnlock()
		return nil
	}

	// It is possible that another channel attempted to create the same mutex and
	// established a lock before this one. Do a final check to see if a lock exists
	if _, exists := nodes[dir]; exists {
		return nil
	}

	// create a mutex for the path
	nodes[dir] = &sync.RWMutex{} // add a rwlock

	node, exists := nodes[dir]
	if !exists {
		return fmt.Errorf(`error creating mutex lock for cloning repo "%v" to dir "%v"`, repo, dir)
	}

	// write lock the path to block others from cloning the same path
	node.Lock() // establish a write lock so the other readers are blocked
	defer node.Unlock()
	coordinator.Unlock()

	// clone the repo
	cmd := exec.Command("git", "clone", repo, dir)
	cmd.Env = append(cmd.Env, os.Environ()...)
	cmd.Env = append(cmd.Env, "GIT_TERMINAL_PROMPT=0")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf(`error cloning git repository "%v" into "%v": %w`, repo, dir, err)
	}

	return nil
}

// checkout will perform a `git checkout <target>` for the git repository found
// at `repo`
func checkout(repo string, target string) error {
	coordinator := pathCoordinator.coordinator
	nodes := pathCoordinator.nodes
	coordinator.Lock()
	if _, exists := nodes[target]; !exists {
		nodes[target] = &sync.RWMutex{}
	}
	node := nodes[target]
	node.Lock()
	defer node.Unlock()
	coordinator.Unlock()

	cmd := exec.Command("git", "checkout", target)
	cmd.Dir = repo
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf(`unable to checkout "%v" from in repository "%v": %w`, target, repo, err)
	}

	return nil
}
