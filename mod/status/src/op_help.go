package status

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/shell"
	"github.com/cryptopunkscc/astrald/mod/status"
)

func (mod *Module) opHelp(ctx astral.Context, env *shell.Env, args opVisibleArgs) (err error) {
	env.Printf("usage: %v <command>\n\n", status.ModuleName)
	env.Printf("commands:\n")
	env.Printf("  scan                          broadcast a scan message to collect statuses\n")
	env.Printf("  show                          show cached statuses\n")
	env.Printf("  update                        broadcast a status update\n")
	env.Printf("  visible [bool]                show or set visibility\n")
	env.Printf("  help                          show help\n")
	
	return nil
}
