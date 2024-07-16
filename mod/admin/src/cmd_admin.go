package admin

import (
	"errors"
	"github.com/cryptopunkscc/astrald/id"
	"github.com/cryptopunkscc/astrald/mod/admin"
)

var _ admin.Command = &CmdAdmin{}

type CmdAdmin struct {
	mod  *Module
	cmds map[string]func(admin.Terminal, []string) error
}

func NewCmdAdmin(mod *Module) *CmdAdmin {
	cmd := &CmdAdmin{mod: mod}
	cmd.cmds = map[string]func(admin.Terminal, []string) error{
		"list":   cmd.list,
		"add":    cmd.add,
		"remove": cmd.remove,
		"help":   cmd.help,
	}
	return cmd
}

func (cmd *CmdAdmin) Exec(term admin.Terminal, args []string) error {
	if len(args) < 2 {
		return cmd.help(term, []string{})
	}

	c, args := args[1], args[2:]
	if fn, found := cmd.cmds[c]; found {
		return fn(term, args)
	}

	return errors.New("unknown command")
}

func (cmd *CmdAdmin) list(term admin.Terminal, _ []string) error {
	var list = cmd.mod.admins.Clone()

	if len(list) == 0 {
		term.Printf("no admins added")
		return nil
	}

	term.Printf("%d admin(s):\n", len(list))

	for _, hex := range list {
		adminID, err := id.ParsePublicKeyHex(hex)
		if err != nil {
			return err
		}
		term.Printf("%v\n", adminID)
	}

	return nil
}

func (cmd *CmdAdmin) add(term admin.Terminal, args []string) error {
	if len(args) < 1 {
		return errors.New("missing argument")
	}

	identity, err := cmd.mod.node.Resolver().Resolve(args[0])
	if err != nil {
		return err
	}

	return cmd.mod.AddAdmin(identity)
}

func (cmd *CmdAdmin) remove(term admin.Terminal, args []string) error {
	if len(args) < 1 {
		return errors.New("missing argument")
	}

	identity, err := cmd.mod.node.Resolver().Resolve(args[0])
	if err != nil {
		return err
	}

	return cmd.mod.RemoveAdmin(identity)
}

func (cmd *CmdAdmin) help(term admin.Terminal, _ []string) error {
	term.Printf("help: %s <command> [options]\n\n", admin.ModuleName)
	term.Printf("commands:\n")
	term.Printf("  list                         list all identities with admin privileges\n")
	term.Printf("  add <identity>               add an admin")
	term.Printf("  remove <identity>            remove an admin")
	term.Printf("  help                         show help\n")
	return nil
}

func (cmd *CmdAdmin) ShortDescription() string {
	return "manage the admin console"
}
