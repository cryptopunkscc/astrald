package nodes

import (
	"context"
	"errors"
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
	streams := adm.mod.streams.Clone()

	slices.SortFunc(streams, func(a, b *Stream) int {
		return a.createdAt.Compare(b.createdAt)
	})

	for _, s := range streams {
		term.Printf("%-4d %-20s %v\n", s.id, s.RemoteIdentity(), s)
	}

	return nil
}

func (adm *Admin) link(term admin.Terminal, args []string) error {
	if len(args) < 1 {
		return errors.New("missing argument")
	}

	remoteID, err := adm.mod.dir.Resolve(args[0])
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	return adm.mod.ensureConnected(ctx, remoteID)
}

func (adm *Admin) conns(term admin.Terminal, args []string) error {
	conns := adm.mod.conns.Values()

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

		term.Printf(
			"%v %-20s %-8s %v\n",
			c.Nonce,
			c.RemoteIdentity,
			state,
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

func (adm *Admin) check(term admin.Terminal, args []string) error {
	if len(args) < 1 {
		return errors.New("not enough arguments")
	}

	identity, err := adm.mod.dir.Resolve(args[0])
	if err != nil {
		return err
	}

	for _, s := range adm.mod.streams.Select(func(s *Stream) bool {
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

	identity, err := adm.mod.dir.Resolve(args[0])
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

func (adm *Admin) ping(term admin.Terminal, args []string) error {
	for _, s := range adm.mod.streams.Clone() {
		term.Printf("%v... ", s.RemoteIdentity())
		rtt := s.Ping()
		term.Printf("%v\n", rtt)
	}

	return nil
}

func (adm *Admin) endpoints(term admin.Terminal, args []string) error {
	if len(args) < 1 {
		return errors.New("not enough arguments")
	}

	identity, err := adm.mod.dir.Resolve(args[0])
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

	identity, err := adm.mod.dir.Resolve(args[0])
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

	identity, err := adm.mod.dir.Resolve(args[0])
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
