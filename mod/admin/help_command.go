package admin

import (
	"sort"
	"strings"
)

var _ Command = &HelpCommand{}

type HelpCommand struct {
	mod *Module
}

type Describer interface {
	HelpDescription() string
}

func (cmd *HelpCommand) Exec(t *Terminal, args []string) error {
	t.Printf("commands:\n\n")

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
		if d, ok := c.(Describer); ok {
			desc = d.HelpDescription()
		}
		t.Printf("  %-12s %s\n", name, desc)
	}

	// legacy commands
	t.Printf("\nlegacy: ")
	var legacy = make([]string, 0, len(commands))
	for c := range commands {
		legacy = append(legacy, c)
	}
	t.Printf("%s\n", strings.Join(legacy, ", "))

	return nil
}

func (cmd *HelpCommand) HelpDescription() string {
	return "show help"
}
