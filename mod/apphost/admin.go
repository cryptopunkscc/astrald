package apphost

import (
	"errors"
	"flag"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/mod/admin/api"
	"os"
	"path/filepath"
	"strconv"
)

type Admin struct {
	mod *Module
}

func (adm *Admin) Exec(t admin.Terminal, args []string) error {
	if len(args) <= 1 {
		return adm.help(t)
	}

	switch args[1] {
	case "run":
		return adm.run(t, args[2:])

	case "list":
		return adm.list(t, args[2:])

	case "kill":
		return adm.kill(t, args[2:])

	case "keys":
		return adm.keys(t, args[2:])

	case "help":
		return adm.help(t)

	default:
		return errors.New("unknown command")
	}
}

func (adm *Admin) ShortDescription() string {
	return "manage apps and permissions"
}

func (adm *Admin) help(out admin.Terminal) error {
	out.Println("usage: apphost <command>")
	out.Println()
	out.Println("commands:")
	out.Println("  run       run an executable with node's identity")
	out.Println("  list      list processes")
	out.Println("  kill      kill a process")
	out.Println("  help      show help")

	return nil
}

func (adm *Admin) run(term admin.Terminal, args []string) error {
	var err error
	var identity = adm.mod.node.Identity()
	var name string
	var f = flag.NewFlagSet("apphost run", flag.ContinueOnError)
	f.SetOutput(term)
	f.StringVar(&name, "i", "", "set identity")
	if err := f.Parse(args); err != nil {
		return err
	}

	args = f.Args()

	if len(args) < 1 {
		f.Usage()
		return errors.New("missing arguments")
	}

	if name != "" {
		identity, err = adm.mod.node.Resolver().Resolve(name)
		if err != nil {
			return err
		}
	}

	_, err = adm.mod.Exec(identity, args[0], args[1:], os.Environ())

	return err
}

func (adm *Admin) list(out admin.Terminal, args []string) error {
	out.Printf("%-6s %-10s %-30s %s\n", "ID", "STATE", "NAME", "IDENTITY")

	for i, e := range adm.mod.execs {
		var identity = adm.mod.node.Resolver().DisplayName(e.identity)
		var name = filepath.Base(e.path)

		out.Printf("%-6d %-10s %-30s %s\n", i, e.State(), name, identity)
	}
	return nil
}

func (adm *Admin) kill(out admin.Terminal, args []string) error {
	if len(args) < 1 {
		out.Println("usage: apphost kill <ID>")
		return errors.New("missing arguments")
	}

	i, err := strconv.Atoi(args[0])
	if err != nil {
		return err
	}

	if (i < 0) || (i >= len(adm.mod.execs)) {
		return errors.New("invalid id")
	}

	return adm.mod.execs[i].Kill()
}

func (adm *Admin) keys(term admin.Terminal, args []string) error {
	if len(args) < 1 {
		return errors.New("keys: missing argument")
	}

	switch args[0] {
	case "new":
		key, err := id.GenerateIdentity()
		if err != nil {
			return err
		}

		if err = adm.mod.keys.Save(key); err != nil {
			return err
		}

		if len(args) >= 2 {
			alias := args[1]
			if err := adm.mod.node.Tracker().SetAlias(key, alias); err != nil {
				term.Printf("cannot set alias: %v\n", err)
			}
		}

		term.Printf("created identity %s (%s)\n", key, admin.Faded(key.String()))
	}

	return nil
}
