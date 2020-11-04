package url

import (
	"net/url"
	"path"
	"strings"
)

// ToPath converts a url to a path like string.
func ToPath(u string) (string, error) {
	noProtocol, err := removeProtocol(u)
	if err != nil {
		return "", err
	}

	var pathSegments []string
	for _, v := range strings.Split(noProtocol, "/") {
		if v != "" {
			pathSegments = append(pathSegments, v)
		}
	}

	return path.Join(pathSegments...), nil
}

func removeProtocol(repoURL string) (string, error) {
	// Return the original URL if it does not start with a protocol
	if !strings.Contains(repoURL, "://") {
		return repoURL, nil
	}

	// Parse the URL, remove the Scheme and leading "/"
	u, err := url.Parse(repoURL)
	if err != nil {
		return "", err
	}
	u.Scheme = ""

	return strings.TrimLeft(u.String(), "/"), nil
}
