package core

type Generator interface {
	Generate(component *Component) (manifest string, err error)
	Install(component *Component) (err error)
}
