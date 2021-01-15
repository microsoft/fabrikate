package generatable

type Generatable interface {
	Generate() error
	GetGeneratePath() string
	Validate() error
}

const generateDirName = "_generated"
