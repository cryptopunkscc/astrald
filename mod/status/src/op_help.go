package status

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/shell"
	"github.com/cryptopunkscc/astrald/mod/status"
)

func (mod *Module) opHelp(ctx astral.Context, q shell.Query) (err error) {
	if v, _ := q.Extra().Get("interface"); v != "terminal" {
		return q.Reject()
	}
	
	t, err := shell.AcceptTerminal(q)
	if err != nil {
		return err
	}
	defer t.Close()

	t.Printf("usage: %v <command>\n\n", status.ModuleName)
	t.Printf("commands:\n")
	t.Printf("  scan                          broadcast a scan message to collect statuses\n")
	t.Printf("  show                          show cached statuses\n")
	t.Printf("  update                        broadcast a status update\n")
	t.Printf("  visible [bool]                show or set visibility\n")
	t.Printf("  help                          show help\n")

	return
}
