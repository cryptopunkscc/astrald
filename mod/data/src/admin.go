package data

import (
	"encoding/json"
	"errors"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"time"
)

type Admin struct {
	mod  *Module
	cmds map[string]func(admin.Terminal, []string) error
}

func NewAdmin(mod *Module) *Admin {
	var cmd = &Admin{mod: mod}
	cmd.cmds = map[string]func(admin.Terminal, []string) error{
		"list":      cmd.list,
		"index":     cmd.index,
		"describe":  cmd.describe,
		"set_label": cmd.setLabel,
		"get_label": cmd.getLabel,
	}
	return cmd
}

func (cmd *Admin) list(term admin.Terminal, args []string) error {
	list, err := cmd.mod.All(time.Time{})
	if err != nil {
		return err
	}

	var format = "%-64s %-8s %-32s %s\n"
	term.Printf(format, admin.Header("ID"), admin.Header("Header"), admin.Header("Type"), admin.Header("Label"))
	for _, item := range list {
		term.Printf(format,
			item.DataID,
			item.Header,
			item.Type,
			cmd.mod.GetLabel(item.DataID),
		)
	}

	return nil
}

func (cmd *Admin) index(term admin.Terminal, args []string) error {
	if len(args) < 1 {
		return errors.New("missing argument")
	}

	dataID, err := data.Parse(args[0])
	if err != nil {
		return err
	}

	return cmd.mod.Index(dataID)
}

func (cmd *Admin) describe(term admin.Terminal, args []string) error {
	if len(args) < 1 {
		return errors.New("missing argument")
	}

	dataID, err := data.Parse(args[0])
	if err != nil {
		return err
	}

	var desc = cmd.mod.DescribeData(nil, dataID, nil)
	bytes, err := json.MarshalIndent(desc, "", "  ")
	if err != nil {
		return err
	}

	term.Write(bytes)
	term.Println()

	return nil
}

func (cmd *Admin) setLabel(term admin.Terminal, args []string) error {
	if len(args) < 2 {
		return errors.New("missing argument")
	}

	dataID, err := data.Parse(args[0])
	if err != nil {
		return err
	}

	label := args[1]

	cmd.mod.SetLabel(dataID, label)

	return nil
}

func (cmd *Admin) getLabel(term admin.Terminal, args []string) error {
	if len(args) < 1 {
		return errors.New("missing argument")
	}

	dataID, err := data.Parse(args[0])
	if err != nil {
		return err
	}

	term.Printf("%s\n", cmd.mod.GetLabel(dataID))
	return nil
}

func (cmd *Admin) Exec(term admin.Terminal, args []string) error {
	if len(args) < 2 {
		return cmd.help(term, []string{})
	}

	c, args := args[1], args[2:]
	if fn, found := cmd.cmds[c]; found {
		return fn(term, args)
	}

	return errors.New("unknown command")
}

func (cmd *Admin) help(term admin.Terminal, _ []string) error {
	term.Printf("usage: data <list|describe|set_label|get_label>\n")
	return nil
}

func (cmd *Admin) ShortDescription() string {
	return "data"
}
