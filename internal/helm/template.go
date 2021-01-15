package helm

import (
	"bytes"
	"fmt"
	"os/exec"
)

// TemplateOptions encapsulate the options for `helm template`
type TemplateOptions struct {
	Release   string
	RepoURL   string
	Chart     string
	Version   string
	Namespace string
	Values    []string
}

// Template is a command for `helm template`
func Template(opts TemplateOptions) (string, error) {
	templateArgs := []string{"template"}
	if opts.Release != "" {
		templateArgs = append(templateArgs, opts.Release)
	}
	templateArgs = append(templateArgs, opts.Chart)
	if opts.RepoURL != "" {
		templateArgs = append(templateArgs, "--repo", opts.RepoURL)
	}
	if opts.Namespace != "" {
		templateArgs = append(templateArgs, "--create-namespace", "--namespace", opts.Namespace)
	}
	for _, yamlPath := range opts.Values {
		templateArgs = append(templateArgs, "--values", yamlPath)
	}

	templateCmd := exec.Command("helm", templateArgs...)
	var stdout, stderr bytes.Buffer
	templateCmd.Stdout = &stdout
	templateCmd.Stderr = &stderr

	if err := templateCmd.Run(); err != nil {
		return stdout.String(), fmt.Errorf("%v: %v", err, stderr.String())
	}

	return stdout.String(), nil
}
