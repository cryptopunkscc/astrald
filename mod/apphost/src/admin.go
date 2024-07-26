package apphost

import (
	"errors"
	"flag"
	"github.com/cryptopunkscc/astrald/id"
	"github.com/cryptopunkscc/astrald/mod/admin"
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
	case "tokens":
		return adm.tokens(t, args[2:])

	case "newtoken":
		return adm.newtoken(t, args[2:])

	case "run":
		return adm.run(t, args[2:])

	case "list":
		return adm.ps(t, args[2:])

	case "kill":
		return adm.kill(t, args[2:])

	case "help":
		return adm.help(t)

	default:
		return errors.New("unknown command")
	}
}

func (adm *Admin) newtoken(term admin.Terminal, args []string) error {
	if len(args) < 1 {
		return errors.New("argument missing: identity")
	}

	identity, err := adm.mod.Dir.Resolve(args[0])
	if err != nil {
		return err
	}

	token, err := adm.mod.CreateAccessToken(identity)
	if err != nil {
		return err
	}

	term.Printf("New access token: %v\n", token)

	return nil
}

func (adm *Admin) tokens(term admin.Terminal, args []string) error {
	var rows []dbAccessToken

	adm.mod.db.Find(&rows)

	const f = "%-34s %v\n"

	term.Printf(f, admin.Header("Token"), admin.Header("Identity"))

	for _, row := range rows {
		identity, err := id.ParsePublicKeyHex(row.Identity)
		if err != nil {
			continue
		}

		term.Printf(f, admin.Keyword(row.Token), identity)
	}

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
		identity, err = adm.mod.Dir.Resolve(name)
		if err != nil {
			return err
		}
	}

	_, err = adm.mod.Exec(identity, args[0], args[1:], os.Environ())

	return err
}

func (adm *Admin) ps(out admin.Terminal, args []string) error {
	out.Printf("%-6s %-10s %-30s %s\n", "ID", "STATE", "NAME", "IDENTITY")

	for i, e := range adm.mod.execs {
		var identity = adm.mod.Dir.DisplayName(e.identity)
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

func (adm *Admin) help(out admin.Terminal) error {
	out.Println("usage: apphost <command>")
	out.Println()
	out.Println("commands:")
	out.Println("  tokens                  list all access tokens")
	out.Println("  newtoken <identity>     create new access token for an identity")
	out.Println("  run                     run an executable")
	out.Println("  ps                      list processes")
	out.Println("  kill                    kill a process")
	out.Println("  help                    show help")

	return nil
}

func (adm *Admin) ShortDescription() string {
	return "manage application host"
}
