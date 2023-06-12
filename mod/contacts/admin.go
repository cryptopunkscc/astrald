package contacts

import (
	"errors"
	"fmt"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node/infra/drivers/gw"
	"github.com/cryptopunkscc/astrald/node/tracker"
	"github.com/cryptopunkscc/astrald/nodeinfo"
	"time"
)

type Admin struct {
	mod  *Module
	cmds map[string]func(*admin.Terminal, []string) error
}

func NewAdmin(mod *Module) *Admin {
	var adm = &Admin{mod: mod}
	adm.cmds = map[string]func(*admin.Terminal, []string) error{
		"add":    adm.add,
		"parse":  adm.parse,
		"list":   adm.list,
		"show":   adm.show,
		"remove": adm.remove,
		"info":   adm.info,
		"help":   adm.help,
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

func (adm *Admin) add(_ *admin.Terminal, args []string) error {
	if len(args) < 1 {
		return errors.New("argument missing")
	}

	info, err := nodeinfo.Parse(args[0])
	if err != nil {
		return err
	}

	if info.Identity.IsEqual(adm.mod.node.Identity()) {
		return errors.New("cannot add self")
	}

	for _, ep := range info.Endpoints {
		ep, err := adm.mod.node.Infra().Unpack(ep.Network(), ep.Pack())
		if err != nil {
			return err
		}
		if err := adm.mod.node.Tracker().AddEndpoint(info.Identity, ep, time.Now().Add(7*24*time.Hour)); err != nil {
			return err
		}
	}

	return adm.mod.node.Tracker().SetAlias(info.Identity, info.Alias)
}

func (adm *Admin) list(term *admin.Terminal, _ []string) error {
	nodes, err := adm.mod.All()
	if err != nil {
		return err
	}

	for _, node := range nodes {
		if err := adm.printNode(term, node, false); err != nil {
			return err
		}
	}

	return nil
}

func (adm *Admin) show(term *admin.Terminal, args []string) error {
	if len(args) < 1 {
		return errors.New("argument missing")
	}

	r := &Resolver{mod: adm.mod}

	identity, err := r.Resolve(args[0])
	if err != nil {
		return err
	}

	node, err := adm.mod.Find(identity)
	if err != nil {
		return err
	}

	if err := adm.printNode(term, node, true); err != nil {
		return err
	}

	return nil
}

func (adm *Admin) printNode(term *admin.Terminal, node Node, showInfo bool) error {
	var info = &nodeinfo.NodeInfo{
		Identity:  node.Identity,
		Alias:     node.Alias,
		Endpoints: []net.Endpoint{},
	}

	term.Printf("%s (%s)\n", node.Alias, node.Identity)

	endpoints, err := adm.mod.node.Tracker().EndpointsByIdentity(node.Identity)
	if err != nil {
		return err
	}

	for _, e := range endpoints {
		info.Endpoints = append(info.Endpoints, e.Endpoint)
		term.Printf("  %s\n", adm.formatEndpoint(e))
	}
	term.Println()

	if showInfo {
		term.Printf("  nodeinfo %s\n", info)
	}

	return nil
}

func (adm *Admin) parse(term *admin.Terminal, args []string) error {
	if len(args) < 1 {
		return errors.New("argument missing")
	}

	info, err := nodeinfo.Parse(args[0])
	if err != nil {
		return err
	}

	term.Printf("%s (%s)\n", info.Alias, info.Identity)
	for _, a := range info.Endpoints {
		ep, err := adm.mod.node.Infra().Unpack(a.Network(), a.Pack())
		if err != nil {
			continue
		}

		term.Printf("  %s\n", adm.formatEndpoint(ep))
	}

	return nil
}

func (adm *Admin) remove(_ *admin.Terminal, args []string) error {
	if len(args) < 1 {
		return errors.New("argument missing")
	}

	var r = &Resolver{mod: adm.mod}

	identity, err := r.Resolve(args[0])
	if err != nil {
		return err
	}

	return adm.mod.Delete(identity)
}

func (adm *Admin) info(term *admin.Terminal, _ []string) error {
	var info = nodeinfo.FromNode(adm.mod.node)

	term.Printf("%s (%s)\n", info.Alias, info.Identity)
	for _, ep := range info.Endpoints {
		term.Printf("  %s\n", adm.formatEndpoint(ep))
	}

	term.Printf("\nnodeinfo %s\n", info)
	return nil
}

func (adm *Admin) help(term *admin.Terminal, _ []string) error {
	term.Printf("usage: contacts <command>\n\n")
	term.Printf("commands:\n")
	term.Printf("  list                           list all contacts\n")
	term.Printf("  add <nodeInfo>                 add contact from nodeinfo\n")
	term.Printf("  set_alias <nodeID> <alias>     set node's alias\n")
	term.Printf("  remove <nodeID>                remove node from contacts\n")
	term.Printf("  show <nodeID>                  show contact information\n")
	term.Printf("  parse <nodeInfo>               parse node info\n")
	term.Printf("  info                           show localnode's info\n")
	term.Printf("  help                           show help\n")
	return nil
}

func (adm *Admin) ShortDescription() string {
	return "manage contacts"
}

func (adm *Admin) formatEndpoint(endpoint net.Endpoint) string {
	var suffix string

	if e, ok := endpoint.(tracker.TrackedEndpoint); ok {
		d := e.ExpiresAt.Sub(time.Now()).Round(time.Second)
		suffix = fmt.Sprintf(" (expires in %s)", d)
		endpoint = e.Endpoint
	}

	network, address := endpoint.Network(), endpoint.String()

	if e, ok := endpoint.(gw.Endpoint); ok {
		var r = adm.mod.node.Resolver()
		address = fmt.Sprintf("%s:%s", r.DisplayName(e.Gate()), r.DisplayName(e.Target()))
	}

	return fmt.Sprintf("%-10s%s"+suffix, network, address)
}
