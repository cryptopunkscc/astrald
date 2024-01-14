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
		"add":    adm.add,
		"remove": adm.remove,
		"list":   adm.list,
		"help":   adm.help,
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

func (adm *Admin) add(term admin.Terminal, args []string) error {
	if len(args) < 1 {
		return errors.New("missing argument")
	}

	identity, err := adm.mod.node.Resolver().Resolve(args[0])
	if err != nil {
		return err
	}

	return adm.mod.AddIdentity(identity)
}

func (adm *Admin) remove(term admin.Terminal, args []string) error {
	if len(args) < 1 {
		return errors.New("missing argument")
	}

	identity, err := adm.mod.node.Resolver().Resolve(args[0])
	if err != nil {
		return err
	}

	return adm.mod.RemoveIdentity(identity)
}

func (adm *Admin) list(term admin.Terminal, args []string) error {
	term.Printf("User identities:\n")

	for _, user := range adm.mod.identities.Clone() {
		var p admin.Important
		if user.identity.PrivateKey() != nil {
			p = " (private key)"
		}
		term.Printf("%s%s\n", user.identity, p)
	}

	return nil
}

func (adm *Admin) ShortDescription() string {
	return "manage user"
}

func (adm *Admin) help(term admin.Terminal, _ []string) error {
	term.Printf("usage: %s <command>\n\n", user.ModuleName)
	term.Printf("commands:\n")
	term.Printf("  add <identity>       add identity to the user\n")
	term.Printf("  remove <identity>    remove identity from the user\n")
	term.Printf("  list                 list user's identities\n")
	term.Printf("  help                 show help\n")
	return nil
}
