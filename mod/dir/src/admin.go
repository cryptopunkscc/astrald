package dir

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/lib/desc"
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
		"describe": adm.describe,
	}

	return adm
}

func (adm *Admin) describe(term admin.Terminal, args []string) error {
	var err error
	var opts = &desc.Opts{
		IdentityFilter: id.AllowEveryone,
	}

	var flags = flag.NewFlagSet("describe", flag.ContinueOnError)
	flags.BoolVar(&opts.Network, "n", false, "use network sources")
	flags.SetOutput(term)
	err = flags.Parse(args)
	if err != nil {
		return err
	}

	if len(flags.Args()) == 0 {
		return errors.New("missing data id")
	}

	identity, err := adm.mod.node.Resolver().Resolve(args[0])
	if err != nil {
		return err
	}

	var descs = adm.mod.Describe(context.Background(), identity, opts)

	for _, d := range descs {
		term.Printf("%v: %v\n  ", d.Source, admin.Keyword(d.Data.Type()))

		j, err := json.MarshalIndent(d.Data, "  ", "  ")
		if err != nil {
			term.Printf("marshal error: %v\n", err)
		}
		term.Printf("%s\n\n", string(j))
	}

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
	term.Printf("usage: %s <command>\n\n", dir.ModuleName)
	term.Printf("commands:\n")
	term.Printf("  help            show help\n")
	return nil
}
