package dir

import (
	"errors"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/mod/dir"
)

type Admin struct {
	mod  *Module
	cmds map[string]func(admin.Terminal, []string) error
}

func NewAdmin(mod *Module) *Admin {
	var adm = &Admin{mod: mod}
	adm.cmds = map[string]func(admin.Terminal, []string) error{
		"help":     adm.help,
		"setalias": adm.setAlias,
		"getalias": adm.getAlias,
		"resolve":  adm.resolve,
	}

	return adm
}

func (adm *Admin) setAlias(term admin.Terminal, args []string) error {
	if len(args) < 2 {
		return errors.New("not enough arguments")
	}

	identity, err := adm.mod.ResolveIdentity(args[0])
	if err != nil {
		return err
	}

	return adm.mod.SetAlias(identity, args[1])
}

func (adm *Admin) getAlias(term admin.Terminal, args []string) error {
	if len(args) < 1 {
		return errors.New("not enough arguments")
	}

	identity, err := adm.mod.ResolveIdentity(args[0])
	if err != nil {
		return err
	}

	alias, err := adm.mod.GetAlias(identity)
	if err != nil {
		return err
	}
	if alias == "" {
		term.Printf("no alias set\n")
	} else {
		term.Printf("%v\n", alias)
	}

	return nil
}

func (adm *Admin) resolve(term admin.Terminal, args []string) error {
	if len(args) < 1 {
		return errors.New("not enough arguments")
	}

	identity, err := adm.mod.ResolveIdentity(args[0])
	if err != nil {
		return err
	}

	term.Printf("%v\n", identity.String())

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
	return "manage " + dir.ModuleName
}

func (adm *Admin) help(term admin.Terminal, _ []string) error {
	term.Printf("usage: %v <command>\n\n", dir.ModuleName)
	term.Printf("commands:\n")
	term.Printf("  help            show help\n")
	return nil
}
