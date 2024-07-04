package nodes

import (
	"context"
	"errors"
	"flag"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/net"
	"time"
)

const defaultLinkTimeout = time.Minute

type Admin struct {
	mod  *Module
	cmds map[string]func(admin.Terminal, []string) error
}

func NewAdmin(mod *Module) *Admin {
	var adm = &Admin{mod: mod}
	adm.cmds = map[string]func(admin.Terminal, []string) error{
		"link": adm.link,
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

func (adm *Admin) link(term admin.Terminal, args []string) error {
	flags := flag.NewFlagSet("net link <nodeID>", flag.ContinueOnError)
	flags.SetOutput(term)
	flags.Usage = func() {
		term.Printf("Usage:\n\n  net link [options] <nodeID>\n\nOptions:\n")
		flags.PrintDefaults()
	}
	var network = flags.String("n", "", "link via this network only")
	var timeout = flags.Duration("t", defaultLinkTimeout, "set timeout")
	var addr = flags.String("a", "", "link via this address (requires -n)")
	err := flags.Parse(args)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	args = flags.Args()

	if len(args) < 1 {
		flags.Usage()
		return nil
	}

	var endpoints []net.Endpoint

	remoteID, err := adm.mod.node.Resolver().Resolve(args[0])
	if err != nil {
		return err
	}

	if *addr != "" {
		if *network == "" {
			return errors.New("linking via address requires specifying the network")
		}
		e, err := adm.mod.node.Infra().Parse(*network, *addr)
		if err != nil {
			return err
		}
		endpoints = []net.Endpoint{e}
	} else {
		endpoints, err = adm.mod.node.Tracker().EndpointsByIdentity(remoteID)
		if err != nil {
			return err
		}

		if *network != "" {
			endpoints = selectEndpoints(endpoints, func(e net.Endpoint) bool {
				return e.Network() == *network
			})
		}
	}

	if len(endpoints) == 0 {
		return errors.New("no usable endpoints")
	}

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	lnk, err := adm.mod.Link(ctx, remoteID, nodes.LinkOpts{Endpoints: endpoints})
	if err != nil {
		return err
	}

	term.Printf("linked via %s\n", net.Network(lnk))

	return nil
}

func (adm *Admin) ShortDescription() string {
	return "manage " + nodes.ModuleName
}

func (adm *Admin) help(term admin.Terminal, _ []string) error {
	term.Printf("usage: %s <command>\n\n", nodes.ModuleName)
	term.Printf("commands:\n")
	term.Printf("  link <identity>     create a new link to a node\n")
	term.Printf("  help                show help\n")
	return nil
}

func selectEndpoints(list []net.Endpoint, selector func(net.Endpoint) bool) []net.Endpoint {
	var filtered = make([]net.Endpoint, 0)
	for _, e := range list {
		if selector(e) {
			filtered = append(filtered, e)
		}
	}
	return filtered
}
