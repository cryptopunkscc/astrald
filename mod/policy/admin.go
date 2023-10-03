package policy

import (
	"github.com/cryptopunkscc/astrald/mod/admin"
)

type Admin struct {
	mod *Module
}

func NewAdmin(mod *Module) *Admin {
	return &Admin{mod: mod}
}

func (adm *Admin) Exec(term *admin.Terminal, args []string) error {
	if len(args) < 2 {
		return adm.help(term, []string{})
	}

	switch args[1] {
	case "add":
		return adm.add(term, args[2:])
	case "list":
		return adm.list(term, args[2:])
	default:
		return adm.help(term, args[2:])
	}
}

func (adm *Admin) list(term *admin.Terminal, args []string) error {
	f := "%s\n"
	term.Printf(f, admin.Header("Name"))
	for policy := range adm.mod.policies {
		term.Printf(
			f,
			policy.Name(),
		)
	}
	return nil
}

func (adm *Admin) help(term *admin.Terminal, _ []string) error {
	term.Printf("usage: policy <command>\n\n")
	term.Printf("commands:\n")
	term.Printf("  list      list running policies\n")
	term.Printf("  add       add a policy\n")
	return nil
}

func (adm *Admin) ShortDescription() string {
	return "manage policies"
}
