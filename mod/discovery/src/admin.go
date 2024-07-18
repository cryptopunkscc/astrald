package discovery

import (
	"context"
	"errors"
	"flag"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/mod/discovery"
	"github.com/cryptopunkscc/astrald/net"
	"reflect"
	"time"
)

type Admin struct {
	mod  *Module
	cmds map[string]func(admin.Terminal, []string) error
}

func NewAdmin(mod *Module) *Admin {
	var adm = &Admin{mod: mod}
	adm.cmds = map[string]func(admin.Terminal, []string) error{
		"help":    adm.help,
		"query":   adm.query,
		"sources": adm.sources,
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

func (adm *Admin) query(term admin.Terminal, args []string) error {
	var err error
	var origin = net.OriginLocal
	var caller = adm.mod.node.Identity()
	var callerArg string
	var target = adm.mod.node.Identity()
	var targetArg string
	var f = flag.NewFlagSet("discovery local", flag.ContinueOnError)
	f.SetOutput(term)
	f.StringVar(&callerArg, "c", "", "set caller identity")
	f.StringVar(&targetArg, "t", "", "set target identity")
	f.StringVar(&origin, "o", net.OriginLocal, "set origin for the query (local/network)")
	f.Parse(args)
	args = f.Args()

	if callerArg != "" {
		caller, err = adm.mod.dir.Resolve(callerArg)
		if err != nil {
			return err
		}
	}

	if targetArg != "" {
		target, err = adm.mod.dir.Resolve(targetArg)
		if err != nil {
			return err
		}
	}

	var info *discovery.Info

	term.Printf("Querying %s as %s...\n", target, caller)

	qctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	if target.IsEqual(adm.mod.node.Identity()) {
		info, err = adm.mod.DiscoverLocal(qctx, caller, origin)
	} else {
		info, err = adm.mod.DiscoverRemote(qctx, target, caller)
	}
	if err != nil {
		return err
	}

	if len(info.Data) == 0 {
		term.Printf("No data discovered.\n")
	} else {
		term.Printf("Discovered %d data items.\n", len(info.Data))
	}

	if len(info.Services) == 0 {
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

	for _, item := range info.Services {
		term.Printf(fmt, item.Identity, admin.Keyword(item.Name), item.Type, log.DataSize(len(item.Extra)))
	}

	return nil
}

func (adm *Admin) sources(term admin.Terminal, _ []string) error {
	var list = adm.mod.services.Clone()

	if len(list) == 0 {
		term.Println("no sources registered")
		return nil
	}

	var f = "%s\n"

	term.Printf(f,
		admin.Header("Source"),
	)

	for src, _ := range list {
		var typ = reflect.TypeOf(src)
		if typ.Kind() == reflect.Ptr {
			typ = typ.Elem()
		}

		term.Printf(f, typ)
	}

	return nil
}

func (adm *Admin) help(term admin.Terminal, _ []string) error {
	term.Printf("usage: %s <command>\n\n", discovery.ModuleName)
	term.Printf("commands:\n")
	var f = "  %-15s %s\n"
	term.Printf(f, "sources", "show registered discovery sources")
	term.Printf(f, "query", "query local/remote services")
	term.Printf(f, "help", "show help")
	return nil
}

func (adm *Admin) ShortDescription() string {
	return "service discovery tools"
}
