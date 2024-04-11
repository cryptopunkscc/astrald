package user

import (
	"errors"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/mod/user"
)

type Admin struct {
	mod  *Module
	cmds map[string]func(admin.Terminal, []string) error
}

func NewAdmin(mod *Module) *Admin {
	var adm = &Admin{mod: mod}
	adm.cmds = map[string]func(admin.Terminal, []string) error{
		"set":  adm.set,
		"info": adm.info,
		"help": adm.help,
	}

	return adm
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

func (adm *Admin) set(term admin.Terminal, args []string) error {
	if len(args) < 1 {
		return errors.New("missing argument")
	}

	identity, err := adm.mod.node.Resolver().Resolve(args[0])
	if err != nil {
		return err
	}

	return adm.mod.SetUserID(identity)
}

func (adm *Admin) info(term admin.Terminal, args []string) error {
	userID := adm.mod.UserID()
	if userID.IsZero() {
		return errors.New("no user identity set")
	}

	term.Printf("Identity: %v\n", userID)
	term.Printf("PubKey:   %v\n", userID.PublicKeyHex())

	return nil
}

func (adm *Admin) ShortDescription() string {
	return "manage user"
}

func (adm *Admin) help(term admin.Terminal, _ []string) error {
	term.Printf("usage: %s <command>\n\n", user.ModuleName)
	term.Printf("commands:\n")
	term.Printf("  set <identity>       set local user identity\n")
	term.Printf("  info                 show user info\n")
	term.Printf("  help                 show help\n")
	return nil
}
