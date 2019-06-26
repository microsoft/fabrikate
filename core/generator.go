package core

// The Generator interface defines the interface for generator tools (like Helm or Static)
// to install and generate resource manifests.
type Generator interface {
	Generate(component *Component) (manifest string, err error)
	Install(component *Component, accessTokens map[string]string) (err error)
}
