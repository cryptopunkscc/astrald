package nodes

import (
	"context"
	"errors"
	"flag"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/nodes"
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
		"link":      adm.link,
		"list":      adm.list,
		"links":     adm.links,
		"ep_add":    adm.addEndpoint,
		"ep_rm":     adm.removeEndpoint,
		"add":       adm.add,
		"parse":     adm.parse,
		"show":      adm.show,
		"endpoints": adm.endpoints,
		"help":      adm.help,
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

	var endpoints []exonet.Endpoint

	remoteID, err := adm.mod.node.Resolver().Resolve(args[0])
	if err != nil {
		return err
	}

	if *addr != "" {
		if *network == "" {
			return errors.New("linking via address requires specifying the network")
		}
		e, err := adm.mod.exonet.Parse(*network, *addr)
		if err != nil {
			return err
		}
		endpoints = []exonet.Endpoint{e}
	} else {
		endpoints = adm.mod.Endpoints(remoteID)

		if *network != "" {
			endpoints = selectEndpoints(endpoints, func(e exonet.Endpoint) bool {
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

	term.Printf("linked via %s\n", exonet.Network(lnk))

	return nil
}

func (adm *Admin) list(term admin.Terminal, args []string) error {
	nodes := adm.mod.Nodes()

	var f = "%-30s %s\n"
	term.Printf(f, admin.Header("Alias"), admin.Header("PubKey"))
	for _, nodeID := range nodes {
		term.Printf(f, nodeID, admin.Faded(nodeID.String()))
	}

	return nil
}

func (adm *Admin) links(term admin.Terminal, args []string) error {
	for _, link := range adm.mod.links.Clone() {
		term.Printf("%v\n", link.RemoteIdentity())
	}

	return nil
}

func (adm *Admin) add(_ admin.Terminal, args []string) error {
	if len(args) < 1 {
		return errors.New("argument missing")
	}

	info, err := adm.mod.ParseInfo(args[0])
	if err != nil {
		return err
	}

	if info.Identity.IsEqual(adm.mod.node.Identity()) {
		return errors.New("cannot add self")
	}

	err = adm.mod.AddEndpoint(info.Identity, info.Endpoints...)
	if err != nil {
		return err
	}

	return adm.mod.dir.SetAlias(info.Identity, info.Alias)
}

func (adm *Admin) parse(term admin.Terminal, args []string) error {
	if len(args) < 1 {
		return errors.New("argument missing")
	}

	info, err := adm.mod.ParseInfo(args[0])
	if err != nil {
		return err
	}

	term.Printf("%s %s (%s)\n\n", admin.Header("Identity"), info.Identity, admin.Faded(info.Identity.PublicKeyHex()))

	var f = "%-10s %-40s\n"
	term.Printf(f, admin.Header("Network"), admin.Header("Address"))
	for _, ep := range info.Endpoints {
		ep, err := adm.mod.exonet.Unpack(ep.Network(), ep.Pack())
		if err != nil {
			continue
		}

		term.Printf(f, ep.Network(), ep)
	}
	term.Printf("%d %s\n", len(info.Endpoints), admin.Faded("endpoint(s)."))

	return nil
}

func (adm *Admin) show(term admin.Terminal, args []string) error {
	if len(args) < 1 {
		return errors.New("not enough arguments")
	}

	identity, err := adm.mod.node.Resolver().Resolve(args[0])
	if err != nil {
		return err
	}

	alias, _ := adm.mod.dir.GetAlias(identity)

	term.Printf("%s (%s)\n", identity, admin.Faded(identity.String()))

	// check private key
	if adm.mod.keys != nil {
		if _, err := adm.mod.keys.FindIdentity(identity.PublicKeyHex()); err == nil {
			term.Printf("%s\n", admin.Important("private key available"))
		}
	}

	term.Println()

	// print endpoints
	var endpoints []exonet.Endpoint

	if identity.IsEqual(adm.mod.node.Identity()) {
		endpoints, _ = adm.mod.exonet.Resolve(context.Background(), adm.mod.node.Identity())
	} else {
		endpoints = adm.mod.Endpoints(identity)
	}

	if len(endpoints) == 0 {
		term.Printf("no known endpoints.\n")
	} else {
		var f = "%-10s %-40s\n"
		term.Printf(f, admin.Header("Network"), admin.Header("Address"))
		for _, ep := range endpoints {
			term.Printf(f, ep.Network(), ep)
		}
		term.Printf("%d %s\n\n", len(endpoints), admin.Faded("endpoint(s)."))

		info := nodes.NodeInfo{
			Identity:  identity,
			Alias:     alias,
			Endpoints: endpoints,
		}

		term.Printf("%s %s\n", admin.Header("nodelink"), adm.mod.InfoString(&info))
	}

	return nil
}

func (adm *Admin) endpoints(term admin.Terminal, args []string) error {
	if len(args) < 1 {
		return errors.New("not enough arguments")
	}

	identity, err := adm.mod.node.Resolver().Resolve(args[0])
	if err != nil {
		return err
	}

	// print endpoints
	var endpoints []exonet.Endpoint

	if identity.IsEqual(adm.mod.node.Identity()) {
		endpoints, _ = adm.mod.exonet.Resolve(context.Background(), adm.mod.node.Identity())
	} else {
		endpoints = adm.mod.Endpoints(identity)
	}

	if len(endpoints) == 0 {
		term.Printf("no known endpoints.\n")
	} else {
		var f = "%-10s %-40s\n"
		term.Printf(f, admin.Header("Network"), admin.Header("Address"))
		for _, ep := range endpoints {
			term.Printf(f, ep.Network(), ep)
		}
		term.Printf("%d %s\n\n", len(endpoints), admin.Faded("endpoint(s)."))
	}

	return nil
}

func (adm *Admin) addEndpoint(term admin.Terminal, args []string) error {
	if len(args) < 3 {
		term.Println("usage: nodes ep_add <node> <network> <address>")
		return errors.New("misisng arguments")
	}

	identity, err := adm.mod.node.Resolver().Resolve(args[0])
	if err != nil {
		return err
	}

	ep, err := adm.mod.exonet.Parse(args[1], args[2])
	if err != nil {
		return err
	}

	err = adm.mod.AddEndpoint(identity, ep)
	if err != nil {
		return err
	}

	term.Printf("%s %v added to %s\n", ep.Network(), ep, identity)

	return nil
}

func (adm *Admin) removeEndpoint(term admin.Terminal, args []string) error {
	if len(args) < 3 {
		term.Println("usage: nodes ep_rm <node> <network> <address>")
		return errors.New("misisng arguments")
	}

	identity, err := adm.mod.node.Resolver().Resolve(args[0])
	if err != nil {
		return err
	}

	ep, err := adm.mod.exonet.Parse(args[1], args[2])
	if err != nil {
		return err
	}

	err = adm.mod.RemoveEndpoint(identity, ep)
	if err != nil {
		return err
	}

	term.Printf("%s %v removed from %s\n", ep.Network(), ep, identity)

	return nil
}

func (adm *Admin) ShortDescription() string {
	return "manage " + nodes.ModuleName
}

func (adm *Admin) help(term admin.Terminal, _ []string) error {
	term.Printf("usage: %s <command>\n\n", nodes.ModuleName)
	term.Printf("commands:\n")
	term.Printf("  link                establish a link to a node\n")
	term.Printf("  list                list known nodes\n")
	term.Printf("  ep_add              add an endpoint to a node\n")
	term.Printf("  ep_rm               remove an endpoint from a node\n")
	term.Printf("  show                show all node info\n")
	term.Printf("  endpoints           show all endpoints of a node\n")
	term.Printf("  help                show help\n")
	return nil
}

func selectEndpoints(list []exonet.Endpoint, selector func(exonet.Endpoint) bool) []exonet.Endpoint {
	var filtered = make([]exonet.Endpoint, 0)
	for _, e := range list {
		if selector(e) {
			filtered = append(filtered, e)
		}
	}
	return filtered
}
