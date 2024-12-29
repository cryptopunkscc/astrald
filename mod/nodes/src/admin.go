package nodes

import (
	"context"
	"errors"
	"flag"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"slices"
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
		"ping":      adm.ping,
		"conns":     adm.conns,
		"check":     adm.check,
		"streams":   adm.streams,
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

func (adm *Admin) streams(term admin.Terminal, args []string) error {
	streams := adm.mod.peers.streams.Clone()

	slices.SortFunc(streams, func(a, b *Stream) int {
		return a.createdAt.Compare(b.createdAt)
	})

	for _, s := range streams {
		var d = "<"
		if s.outbound {
			d = ">"
		}

		term.Printf("%v %v %v %v\n", s.id, d, s.RemoteIdentity(), s)
	}

	return nil
}

func (adm *Admin) link(term admin.Terminal, args []string) (err error) {
	if len(args) < 1 {
		return errors.New("missing argument")
	}

	var onlyNet string

	var f = flag.NewFlagSet("link", flag.ContinueOnError)
	f.StringVar(&onlyNet, "n", "", "connect via this network only")
	err = f.Parse(args)
	if err != nil {
		return err
	}

	args = f.Args()

	remoteID, err := adm.mod.Dir.ResolveIdentity(args[0])
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	endpoints := adm.mod.Endpoints(remoteID)
	if len(onlyNet) > 0 {
		endpoints = slices.DeleteFunc(endpoints, func(e exonet.Endpoint) bool {
			return e.Network() != onlyNet
		})
	}

	return adm.mod.peers.connectAny(ctx, remoteID, endpoints)
}

func (adm *Admin) conns(term admin.Terminal, args []string) error {
	conns := adm.mod.peers.conns.Values()

	slices.SortFunc(conns, func(a, b *conn) int {
		return a.createdAt.Compare(b.createdAt)
	})

	for _, c := range conns {
		var state = "?"
		switch c.state.Load() {
		case 0:
			state = "routing"
		case 1:
			state = "open"
		case 2:
			state = "closed"
			continue
		}

		var d = "<"
		if c.Outbound {
			d = ">"
		}

		term.Printf(
			"%v %v %v %v %v/%v %v %v\n",
			c.Nonce,
			d,
			c.RemoteIdentity,
			state,
			c.rused,
			c.rsize,
			c.wsize,
			c.Query,
		)
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

	return adm.mod.Dir.SetAlias(info.Identity, info.Alias)
}

func (adm *Admin) parse(term admin.Terminal, args []string) error {
	if len(args) < 1 {
		return errors.New("argument missing")
	}

	info, err := adm.mod.ParseInfo(args[0])
	if err != nil {
		return err
	}

	term.Printf("%v %v (%v)\n\n", admin.Header("Identity"), info.Identity, admin.Faded(info.Identity.String()))

	var f = "%v %v\n"
	term.Printf(f, admin.Header("Network"), admin.Header("Address"))
	for _, ep := range info.Endpoints {
		ep, err := adm.mod.Exonet.Unpack(ep.Network(), ep.Pack())
		if err != nil {
			continue
		}

		term.Printf(f, ep.Network(), ep)
	}
	term.Printf("%v %v\n", len(info.Endpoints), admin.Faded("endpoint(s)."))

	return nil
}

func (adm *Admin) check(term admin.Terminal, args []string) error {
	if len(args) < 1 {
		return errors.New("not enough arguments")
	}

	identity, err := adm.mod.Dir.ResolveIdentity(args[0])
	if err != nil {
		return err
	}

	for _, s := range adm.mod.peers.streams.Select(func(s *Stream) bool {
		return s.RemoteIdentity().IsEqual(identity)
	}) {
		s.check()
	}

	return nil
}

func (adm *Admin) show(term admin.Terminal, args []string) error {
	if len(args) < 1 {
		return errors.New("not enough arguments")
	}

	identity, err := adm.mod.Dir.ResolveIdentity(args[0])
	if err != nil {
		return err
	}

	alias, _ := adm.mod.Dir.GetAlias(identity)

	term.Printf("%v (%v)\n", identity, admin.Faded(identity.String()))

	// check private key
	if adm.mod.Keys != nil {
		if _, err := adm.mod.Keys.FindIdentity(identity.String()); err == nil {
			term.Printf("%v\n", admin.Important("private key available"))
		}
	}

	term.Println()

	// print endpoints
	var endpoints []exonet.Endpoint

	if identity.IsEqual(adm.mod.node.Identity()) {
		endpoints, _ = adm.mod.Exonet.ResolveEndpoints(context.Background(), adm.mod.node.Identity())
	} else {
		endpoints = adm.mod.Endpoints(identity)
	}

	if len(endpoints) == 0 {
		term.Printf("no known endpoints.\n")
	} else {
		var f = "%v %v\n"
		term.Printf(f, admin.Header("Network"), admin.Header("Address"))
		for _, ep := range endpoints {
			term.Printf(f, ep.Network(), ep)
		}
		term.Printf("%v %v\n\n", len(endpoints), admin.Faded("endpoint(s)."))

		info := nodes.NodeInfo{
			Identity:  identity,
			Alias:     alias,
			Endpoints: endpoints,
		}

		term.Printf("%v %v\n", admin.Header("nodelink"), adm.mod.InfoString(&info))
	}

	return nil
}

func (adm *Admin) ping(term admin.Terminal, args []string) error {
	for _, s := range adm.mod.peers.streams.Clone() {
		term.Printf("%v... ", s.RemoteIdentity())
		rtt, err := s.Ping()
		if err != nil {
			term.Printf("%v\n", err)
		} else {
			term.Printf("%v\n", rtt)
		}
	}

	return nil
}

func (adm *Admin) endpoints(term admin.Terminal, args []string) error {
	if len(args) < 1 {
		return errors.New("not enough arguments")
	}

	identity, err := adm.mod.Dir.ResolveIdentity(args[0])
	if err != nil {
		return err
	}

	// print endpoints
	var endpoints []exonet.Endpoint

	if identity.IsEqual(adm.mod.node.Identity()) {
		endpoints, _ = adm.mod.Exonet.ResolveEndpoints(context.Background(), adm.mod.node.Identity())
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
		term.Printf("%v %v\n\n", len(endpoints), admin.Faded("endpoint(s)."))
	}

	return nil
}

func (adm *Admin) addEndpoint(term admin.Terminal, args []string) error {
	if len(args) < 3 {
		term.Println("usage: nodes ep_add <node> <network> <address>")
		return errors.New("misisng arguments")
	}

	identity, err := adm.mod.Dir.ResolveIdentity(args[0])
	if err != nil {
		return err
	}

	ep, err := adm.mod.Exonet.Parse(args[1], args[2])
	if err != nil {
		return err
	}

	err = adm.mod.AddEndpoint(identity, ep)
	if err != nil {
		return err
	}

	term.Printf("%v %v added to %v\n", ep.Network(), ep, identity)

	return nil
}

func (adm *Admin) removeEndpoint(term admin.Terminal, args []string) error {
	if len(args) < 3 {
		term.Println("usage: nodes ep_rm <node> <network> <address>")
		return errors.New("misisng arguments")
	}

	identity, err := adm.mod.Dir.ResolveIdentity(args[0])
	if err != nil {
		return err
	}

	ep, err := adm.mod.Exonet.Parse(args[1], args[2])
	if err != nil {
		return err
	}

	err = adm.mod.RemoveEndpoint(identity, ep)
	if err != nil {
		return err
	}

	term.Printf("%v %v removed from %v\n", ep.Network(), ep, identity)

	return nil
}

func (adm *Admin) ShortDescription() string {
	return "manage " + nodes.ModuleName
}

func (adm *Admin) help(term admin.Terminal, _ []string) error {
	term.Printf("usage: %v <command>\n\n", nodes.ModuleName)
	term.Printf("commands:\n")
	term.Printf("  link                link to a node\n")
	term.Printf("  streams             show streams\n")
	term.Printf("  conns               list connections\n")
	term.Printf("  ep_add              add an endpoint to a node\n")
	term.Printf("  ep_rm               remove an endpoint from a node\n")
	term.Printf("  show                show all node info\n")
	term.Printf("  endpoints           show all endpoints of a node\n")
	term.Printf("  help                show help\n")
	return nil
}
