package admin

import (
	"github.com/cryptopunkscc/astrald/mod/admin"
	"sort"
)

var _ admin.Command = &CmdHelp{}

type CmdHelp struct {
	mod *Module
}

type ShortDescriber interface {
	ShortDescription() string
}

func (cmd *CmdHelp) Exec(term admin.Terminal, _ []string) error {
	term.Printf("commands:\n\n")

	// get a sorted command list
	var names = make([]string, 0, len(cmd.mod.commands))
	for name := range cmd.mod.commands {
		names = append(names, name)
	}
	sort.Strings(names)

	// display command list and description
	for _, name := range names {
		c := cmd.mod.commands[name]
		var desc string
		if d, ok := c.(ShortDescriber); ok {
			desc = d.ShortDescription()
		}
		term.Printf("  %-12s %s\n", admin.Keyword(name), desc)
	}

	return nil
}

func (cmd *CmdHelp) ShortDescription() string {
	return "show help"
}
