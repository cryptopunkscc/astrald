package shell

import "github.com/cryptopunkscc/astrald/mod/scheduler"

const ModuleName = "shell"

type Module interface {
	Root() *Scope
	NewLogAction(message string) LogAction
}

type LogAction interface {
	scheduler.Task
}
