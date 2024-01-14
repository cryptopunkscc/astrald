package admin

import (
	"errors"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/nodeinfo"
)

var _ admin.Command = &CmdTracker{}

type CmdTracker struct {
	mod  *Module
	cmds map[string]func(admin.Terminal, []string) error
}

func NewCmdTracker(mod *Module) *CmdTracker {
	cmd := &CmdTracker{mod: mod}
	cmd.cmds = map[string]func(admin.Terminal, []string) error{
		"list":         cmd.list,
		"add":          cmd.add,
		"add_endpoint": cmd.addEndpoint,
		"set_alias":    cmd.setAlias,
		"show":         cmd.show,
		"parse":        cmd.parse,
		"clear":        cmd.clear,
		"remove":       cmd.remove,
		"help":         cmd.help,
	}
	return cmd
}

func (cmd *CmdTracker) Exec(term admin.Terminal, args []string) error {
	if len(args) < 2 {
		return cmd.help(term, []string{})
	}

	c, args := args[1], args[2:]
	if fn, found := cmd.cmds[c]; found {
		return fn(term, args)
	}

	return errors.New("unknown command")
}

func (cmd *CmdTracker) list(term admin.Terminal, _ []string) error {
	ids, err := cmd.mod.node.Tracker().Identities()
	if err != nil {
		return err
	}

	var f = "%-30s %s\n"
	term.Printf(f, admin.Header("Alias"), admin.Header("PubKey"))
	for _, nodeID := range ids {
		term.Printf(f, nodeID, admin.Faded(nodeID.String()))
	}

	return nil
}

func (cmd *CmdTracker) show(term admin.Terminal, args []string) error {
	if len(args) < 1 {
		return errors.New("not enough arguments")
	}

	identity, err := cmd.mod.node.Resolver().Resolve(args[0])
	if err != nil {
		return err
	}

	alias, _ := cmd.mod.node.Tracker().GetAlias(identity)

	term.Printf("%s (%s)\n", identity, admin.Faded(identity.String()))

	// check private key
	if cmd.mod.keys != nil {
		if _, err := cmd.mod.keys.FindIdentity(identity.PublicKeyHex()); err == nil {
			term.Printf("%s\n", admin.Important("private key available"))
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
		term.Printf(f, admin.Header("Network"), admin.Header("Address"))
		for _, ep := range endpoints {
			term.Printf(f, ep.Network(), ep)
		}
		term.Printf("%d %s\n\n", len(endpoints), admin.Faded("endpoint(s)."))

		info := nodeinfo.NodeInfo{
			Identity:  identity,
			Alias:     alias,
			Endpoints: endpoints,
		}
		term.Printf("%s %s\n", admin.Header("nodelink"), info.String())
	}

	return nil
}

func (cmd *CmdTracker) addEndpoint(term admin.Terminal, args []string) error {
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

	err = cmd.mod.node.Tracker().AddEndpoint(identity, ep)
	if err != nil {
		return err
	}

	term.Printf("%s %v added to %s\n", ep.Network(), ep, identity)

	return nil
}

func (cmd *CmdTracker) parse(term admin.Terminal, args []string) error {
	if len(args) < 1 {
		return errors.New("argument missing")
	}

	info, err := nodeinfo.Parse(args[0])
	if err != nil {
		return err
	}

	term.Printf("%s %s (%s)\n\n", admin.Header("Identity"), info.Identity, admin.Faded(info.Identity.PublicKeyHex()))

	var f = "%-10s %-40s\n"
	term.Printf(f, admin.Header("Network"), admin.Header("Address"))
	for _, ep := range info.Endpoints {
		ep, err := cmd.mod.node.Infra().Unpack(ep.Network(), ep.Pack())
		if err != nil {
			continue
		}

		term.Printf(f, ep.Network(), ep)
	}
	term.Printf("%d %s\n", len(info.Endpoints), admin.Faded("endpoint(s)."))

	return nil
}

func (cmd *CmdTracker) add(_ admin.Terminal, args []string) error {
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

	return nodeinfo.SaveToNode(info, cmd.mod.node, true)
}

func (cmd *CmdTracker) setAlias(term admin.Terminal, args []string) error {
	if len(args) < 2 {
		return errors.New("not enough arguments")
	}

	identity, err := cmd.mod.node.Resolver().Resolve(args[0])
	if err != nil {
		return err
	}

	return cmd.mod.node.Tracker().SetAlias(identity, args[1])
}

func (cmd *CmdTracker) clear(term admin.Terminal, args []string) error {
	if len(args) < 1 {
		term.Println("usage: tracker clear <identity>")
		return errors.New("misisng arguments")
	}

	identity, err := cmd.mod.node.Resolver().Resolve(args[0])
	if err != nil {
		return err
	}

	return cmd.mod.node.Tracker().Clear(identity)
}

func (cmd *CmdTracker) remove(term admin.Terminal, args []string) error {
	if len(args) < 1 {
		term.Println("usage: tracker remove <identity>")
		return errors.New("misisng arguments")
	}

	identity, err := cmd.mod.node.Resolver().Resolve(args[0])
	if err != nil {
		return err
	}

	return cmd.mod.node.Tracker().Remove(identity)
}

func (cmd *CmdTracker) help(term admin.Terminal, _ []string) error {
	term.Printf("help: tracker <command> [options]\n\n")
	term.Printf("commands:\n")
	term.Printf("  list                                    list all identities\n")
	term.Printf("  show <identity>                         show detailed info about an identity\n")
	term.Printf("  add_endpoint <identity> <net> <addr>    add an endpoint to an identity\n")
	term.Printf("  parse <nodelink>                        parse nodelink data\n")
	term.Printf("  add <nodelink>                          add nodelink data\n")
	term.Printf("  set_alias <identity> <alias>            set identity's alias\n")
	term.Printf("  clear <identity>                        clear identity's endpoints\n")
	term.Printf("  remove <identity>                       remove all identity data\n")
	term.Printf("  help                                    show help\n")
	return nil
}

func (cmd *CmdTracker) ShortDescription() string {
	return "manage node tracker entries"
}
