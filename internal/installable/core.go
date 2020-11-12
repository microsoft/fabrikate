package installable

type Installable interface {
	Install() error
	GetInstallPath() (string, error)
	Validate() error
}

const installDirName = "_components"
