package setup

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/mod/setup"
)

type Admin struct {
	mod  *Module
	cmds map[string]func(admin.Terminal, []string) error
}

func NewAdmin(mod *Module) *Admin {
	var adm = &Admin{mod: mod}
	adm.cmds = map[string]func(admin.Terminal, []string) error{
		"invite": adm.invite,
		"help":   adm.help,
	}

	return adm
}

func (adm *Admin) invite(term admin.Terminal, args []string) error {
	if len(args) == 0 {
		return errors.New("missing argument")
	}

	if term.UserIdentity().IsEqual(adm.mod.node.Identity()) {
		return errors.New("cannot invite as node")
	}

	nodeID, err := adm.mod.dir.Resolve(args[0])
	if err != nil {
		return err
	}

	if nodeID.IsEqual(adm.mod.node.Identity()) {
		return errors.New("cannot invite self")
	}

	return adm.mod.inviteService.Invite(context.Background(), term.UserIdentity(), nodeID)
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
	return "manage " + setup.ModuleName
}

func (adm *Admin) help(term admin.Terminal, _ []string) error {
	term.Printf("usage: %s <command>\n\n", setup.ModuleName)
	term.Printf("commands:\n")
	term.Printf("  help            show help\n")
	return nil
}
