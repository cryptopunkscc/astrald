package shell

import (
	"github.com/cryptopunkscc/astrald/mod/scheduler"
)

const ModuleName = "shell"

type Module interface {
	NewLogAction(message string) LogAction
}

type LogAction interface {
	scheduler.Task
}
