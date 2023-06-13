package admin

import (
	"errors"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/nodeinfo"
	"time"
)

const defaultAddDuration = 30 * 24 * time.Hour

var _ Command = &CmdTracker{}

type CmdTracker struct {
	mod  *Module
	cmds map[string]func(*Terminal, []string) error
}

func NewCmdTracker(mod *Module) *CmdTracker {
	cmd := &CmdTracker{mod: mod}
	cmd.cmds = map[string]func(*Terminal, []string) error{
		"list":         cmd.list,
		"add":          cmd.add,
		"add_endpoint": cmd.addEndpoint,
		"set_alias":    cmd.setAlias,
		"show":         cmd.show,
		"parse":        cmd.parse,
		"remove":       cmd.remove,
		"help":         cmd.help,
	}
	return cmd
}

func (cmd *CmdTracker) Exec(term *Terminal, args []string) error {
	if len(args) < 2 {
		return cmd.help(term, []string{})
	}

	c, args := args[1], args[2:]
	if fn, found := cmd.cmds[c]; found {
		return fn(term, args)
	}

	return errors.New("unknown command")
}

func (cmd *CmdTracker) list(term *Terminal, _ []string) error {
	ids, err := cmd.mod.node.Tracker().Identities()
	if err != nil {
		return err
	}

	var f = "%-30s %s\n"
	term.Printf(f, Header("Alias"), Header("PubKey"))
	for _, nodeID := range ids {
		term.Printf(f, nodeID, Faded(nodeID.String()))
	}

	return nil
}

func (cmd *CmdTracker) show(term *Terminal, args []string) error {
	if len(args) < 1 {
		return errors.New("not enough arguments")
	}

	identity, err := cmd.mod.node.Resolver().Resolve(args[0])
	if err != nil {
		return err
	}

	alias, _ := cmd.mod.node.Tracker().GetAlias(identity)

	term.Printf("%s (%s)\n", identity, Faded(identity.String()))

	// check private key
	if keys, err := cmd.mod.assets.KeyStore(); err == nil {
		if pi, err := keys.Find(identity); err == nil {
			if pi.PrivateKey() != nil {
				term.Printf("%s\n", Important("private key available"))
			}
		}
	}

	term.Println()

	// print endpoints
	var endpoints []net.Endpoint

	if identity.IsEqual(cmd.mod.node.Identity()) {
		endpoints = cmd.mod.node.Infra().Endpoints()
	} else {
		endpoints, err = cmd.mod.node.Tracker().EndpointsByIdentity(identity)
	}

	if len(endpoints) == 0 {
		term.Printf("no known endpoints.\n")
	} else {
		var f = "%-10s %-40s\n"
		term.Printf(f, Header("Network"), Header("Address"))
		for _, ep := range endpoints {
			term.Printf(f, ep.Network(), ep)
		}
		term.Printf("%d %s\n\n", len(endpoints), Faded("endpoint(s)."))

		info := nodeinfo.NodeInfo{
			Identity:  identity,
			Alias:     alias,
			Endpoints: endpoints,
		}
		term.Printf("%s %s\n", Header("nodelink"), info.String())
	}

	return nil
}

func (cmd *CmdTracker) addEndpoint(term *Terminal, args []string) error {
	if len(args) < 3 {
		term.Println("usage: tracker add_endpoint <node> <network> <address>")
		return errors.New("misisng arguments")
	}

	identity, err := cmd.mod.node.Resolver().Resolve(args[0])
	if err != nil {
		return err
	}

	ep, err := cmd.mod.node.Infra().Parse(args[1], args[2])
	if err != nil {
		return err
	}

	return cmd.mod.node.Tracker().AddEndpoint(identity, ep)
}

func (cmd *CmdTracker) parse(term *Terminal, args []string) error {
	if len(args) < 1 {
		return errors.New("argument missing")
	}

	info, err := nodeinfo.Parse(args[0])
	if err != nil {
		return err
	}

	term.Printf("%s %s (%s)\n\n", Header("Identity"), info.Identity, Faded(info.Identity.PublicKeyHex()))

	var f = "%-10s %-40s\n"
	term.Printf(f, Header("Network"), Header("Address"))
	for _, ep := range info.Endpoints {
		ep, err := cmd.mod.node.Infra().Unpack(ep.Network(), ep.Pack())
		if err != nil {
			continue
		}

		term.Printf(f, ep.Network(), ep)
	}
	term.Printf("%d %s\n", len(info.Endpoints), Faded("endpoint(s)."))

	return nil
}

func (cmd *CmdTracker) add(_ *Terminal, args []string) error {
	if len(args) < 1 {
		return errors.New("argument missing")
	}

	info, err := nodeinfo.Parse(args[0])
	if err != nil {
		return err
	}

	if info.Identity.IsEqual(cmd.mod.node.Identity()) {
		return errors.New("cannot add self")
	}

	for _, ep := range info.Endpoints {
		ep, err := cmd.mod.node.Infra().Unpack(ep.Network(), ep.Pack())
		if err != nil {
			return err
		}
		if err := cmd.mod.node.Tracker().AddEndpoint(info.Identity, ep); err != nil {
			return err
		}
	}

	return cmd.mod.node.Tracker().SetAlias(info.Identity, info.Alias)
}

func (cmd *CmdTracker) setAlias(term *Terminal, args []string) error {
	if len(args) < 2 {
		return errors.New("not enough arguments")
	}

	identity, err := cmd.mod.node.Resolver().Resolve(args[0])
	if err != nil {
		return err
	}

	return cmd.mod.node.Tracker().SetAlias(identity, args[1])
}

func (cmd *CmdTracker) remove(term *Terminal, args []string) error {
	if len(args) < 1 {
		term.Println("usage: tracker remove <identity>")
		return errors.New("misisng arguments")
	}

	identity, err := cmd.mod.node.Resolver().Resolve(args[0])
	if err != nil {
		return err
	}

	return cmd.mod.node.Tracker().DeleteAll(identity)
}

func (cmd *CmdTracker) help(term *Terminal, _ []string) error {
	term.Printf("help: tracker <command> [options]\n\n")
	term.Printf("commands:\n")
	term.Printf("  list                                    list all identities\n")
	term.Printf("  show <identity>                         show detailed info about an identity\n")
	term.Printf("  add_endpoint <identity> <net> <addr>    add an endpoint to an identity\n")
	term.Printf("  parse <nodelink>                        parse nodelink data\n")
	term.Printf("  add <nodelink>                          add nodelink data\n")
	term.Printf("  set_alias <identity> <alias>            set identity's alias\n")
	term.Printf("  remove <identity>                       delete identity's endpoints\n")
	term.Printf("  help                                    show help\n")
	return nil
}

func (cmd *CmdTracker) ShortDescription() string {
	return "manage node tracker entries"
}
