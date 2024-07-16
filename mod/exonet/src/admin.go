package exonet

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/mod/exonet"
)

type Admin struct {
	mod  *Module
	cmds map[string]func(admin.Terminal, []string) error
}

func NewAdmin(mod *Module) *Admin {
	var adm = &Admin{mod: mod}
	adm.cmds = map[string]func(admin.Terminal, []string) error{
		"help":    adm.help,
		"resolve": adm.resolve,
	}

	return adm
}

func (adm *Admin) resolve(term admin.Terminal, args []string) error {
	if len(args) == 0 {
		return errors.New("missing argument")
	}

	identity, err := adm.mod.node.Resolver().Resolve(args[0])
	if err != nil {
		return err
	}

	list, err := adm.mod.Resolve(context.Background(), identity)
	if err != nil {
		return err
	}

	for _, endpoint := range list {
		term.Printf("%-8s %v\n", endpoint.Network(), endpoint)
	}

	return nil
}

func (adm *Admin) Exec(term admin.Terminal, args []string) error {
	if len(args) < 2 {
		return adm.help(term, []string{})
	}

	cmd, args := args[1], args[2:]
	if fn, found := adm.cmds[cmd]; found {
		return fn(term, args)
	}

	return errors.New("unknown command")
}

func (adm *Admin) ShortDescription() string {
	return "manage " + exonet.ModuleName
}

func (adm *Admin) help(term admin.Terminal, _ []string) error {
	term.Printf("usage: %s <command>\n\n", exonet.ModuleName)
	term.Printf("commands:\n")
	term.Printf("  help            show help\n")
	return nil
}
