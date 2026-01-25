package shell

import (
	"github.com/cryptopunkscc/astrald/lib/ops"
	"github.com/cryptopunkscc/astrald/mod/scheduler"
)

const ModuleName = "shell"

type Module interface {
	Root() *ops.Set
	NewLogAction(message string) LogAction
}

type LogAction interface {
	scheduler.Task
}
