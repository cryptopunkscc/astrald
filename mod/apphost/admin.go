package apphost

import (
	"errors"
	"fmt"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"os"
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

	case "apps":
		return adm.apps(t, args[2:])

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
	if len(args) < 1 {
		out.Println("usage: apphost run <appname> [args...]")
		return errors.New("missing arguments")
	}

	return adm.mod.Launch(args[0], args[1:], os.Environ())
}

func (adm *Admin) apps(out *admin.Terminal, args []string) error {
	for name, app := range adm.mod.config.Apps {
		var displayName string
		identity, err := adm.mod.node.Resolver().Resolve(app.Identity)
		if err != nil {
			displayName = app.Identity
		} else {
			displayName = adm.mod.node.Resolver().DisplayName(identity)
		}
		fmt.Fprintf(out, "- %s (as %s): %s %s\n", name, displayName, app.Runtime, app.Path)
	}

	return nil
}
