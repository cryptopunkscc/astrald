package discovery

import (
	"context"
	"errors"
	"flag"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/query"
	"reflect"
)

type Admin struct {
	mod  *Module
	cmds map[string]func(*admin.Terminal, []string) error
}

func NewAdmin(mod *Module) *Admin {
	var adm = &Admin{mod: mod}
	adm.cmds = map[string]func(*admin.Terminal, []string) error{
		"help":    adm.help,
		"query":   adm.query,
		"sources": adm.sources,
	}
	return adm
}

func (adm *Admin) Exec(term *admin.Terminal, args []string) error {
	if len(args) < 2 {
		return adm.help(term, []string{})
	}

	cmd, args := args[1], args[2:]
	if fn, found := adm.cmds[cmd]; found {
		return fn(term, args)
	}

	return errors.New("unknown command")
}

func (adm *Admin) query(term *admin.Terminal, args []string) error {
	var err error
	var origin = query.OriginLocal
	var caller = adm.mod.node.Identity()
	var callerArg string
	var target = adm.mod.node.Identity()
	var targetArg string
	var f = flag.NewFlagSet("discovery local", flag.ContinueOnError)
	f.SetOutput(term)
	f.StringVar(&callerArg, "c", "", "set caller identity")
	f.StringVar(&targetArg, "t", "", "set target identity")
	f.StringVar(&origin, "o", query.OriginLocal, "set origin for the query (local/network)")
	f.Parse(args)
	args = f.Args()

	if callerArg != "" {
		caller, err = adm.mod.node.Resolver().Resolve(callerArg)
		if err != nil {
			return err
		}
	}

	if targetArg != "" {
		target, err = adm.mod.node.Resolver().Resolve(targetArg)
		if err != nil {
			return err
		}
	}

	var list []ServiceEntry

	var targetName = adm.mod.node.Resolver().DisplayName(target)
	var callerName = adm.mod.node.Resolver().DisplayName(caller)

	term.Printf("Querying %s as %s...\n", targetName, callerName)

	if target.IsEqual(adm.mod.node.Identity()) {
		list, err = adm.mod.QueryLocal(context.Background(), caller, origin)
	} else {
		list, err = adm.mod.QueryRemoteAs(context.Background(), target, caller)
	}
	if err != nil {
		return err
	}

	if len(list) == 0 {
		term.Printf("No services discovered.\n")
		return nil
	}

	fmt := "%-20s %-30s %-30s %s\n"

	term.Printf(
		fmt,
		admin.Header("Identity"),
		admin.Header("Service"),
		admin.Header("Type"),
		admin.Header("Extra"),
	)

	for _, item := range list {
		term.Printf(fmt, item.Identity, admin.Keyword(item.Name), item.Type, item.Extra)
	}

	return nil
}

func (adm *Admin) sources(term *admin.Terminal, _ []string) error {
	var list = adm.mod.sources

	if len(list) == 0 {
		term.Println("no sources registered")
		return nil
	}

	var f = "%-25s %s\n"

	term.Printf(f,
		admin.Header("Identity"),
		admin.Header("Type"),
	)

	for src, identity := range list {
		var typ = reflect.TypeOf(src).String()
		if s, ok := src.(*ServiceSource); ok {
			typ = "service: " + s.service
		}

		term.Printf(f, identity, typ)
	}

	return nil
}

func (adm *Admin) help(term *admin.Terminal, _ []string) error {
	term.Printf("usage: discovery <command>\n\n")
	term.Printf("commands:\n")
	var f = "  %-15s %s\n"
	term.Printf(f, "sources", "show registered discovery sources")
	term.Printf(f, "query", "query local/remote services")
	term.Printf(f, "help", "show help")
	return nil
}

func (adm *Admin) ShortDescription() string {
	return "manage contacts"
}
