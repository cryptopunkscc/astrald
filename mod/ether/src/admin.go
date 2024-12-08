package ether

import (
	"errors"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/mod/ether"
	"github.com/cryptopunkscc/astrald/object"
)

type Admin struct {
	mod  *Module
	cmds map[string]func(admin.Terminal, []string) error
}

func NewAdmin(mod *Module) *Admin {
	var adm = &Admin{mod: mod}
	adm.cmds = map[string]func(admin.Terminal, []string) error{
		"push": adm.push,
		"help": adm.help,
	}

	return adm
}

func (adm *Admin) push(term admin.Terminal, args []string) error {
	if len(args) == 0 {
		return errors.New("missing argument")
	}

	objectID, err := object.ParseID(args[0])
	if err != nil {
		return err
	}

	obj, err := adm.mod.Objects.Load(objectID)
	if err != nil {
		return err
	}

	return adm.mod.Push(obj, nil)
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
	return "manage " + ether.ModuleName
}

func (adm *Admin) help(term admin.Terminal, _ []string) error {
	term.Printf("usage: %s <command>\n\n", ether.ModuleName)
	term.Printf("commands:\n")
	term.Printf("  push [objectID]   push an object to the ether\n")
	term.Printf("  help              show help\n")
	return nil
}
