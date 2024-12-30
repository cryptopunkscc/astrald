package shell

const ModuleName = "shell"

type Module interface {
	Root() *Scope
}
