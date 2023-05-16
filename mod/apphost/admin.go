package apphost

import (
	"errors"
	"github.com/cryptopunkscc/astrald/mod/admin"
)

type Admin struct {
	mod *Module
}

func (adm *Admin) Exec(t *admin.Terminal, args []string) error {
	if len(args) <= 1 {
		return adm.usage(t)
	}

	switch args[1] {
	case "run":
		return adm.run(t, args[2:])

	case "help":
		return adm.usage(t)

	default:
		return errors.New("unknown command")
	}
}

func (adm *Admin) ShortDescription() string {
	return "manage apps and permissions"
}

func (adm *Admin) usage(out *admin.Terminal) error {
	out.Println("usage: apphost <command>")
	out.Println()
	out.Println("commands: run, help")

	return nil
}

func (adm *Admin) run(out *admin.Terminal, args []string) error {
	if len(args) < 2 {
		out.Println("usage: apphost run <runtime> <app_path>")
		return errors.New("missing arguments")
	}

	return adm.mod.Launch(args[0], args[1])
}
