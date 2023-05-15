package admin

import (
	"errors"
	"fmt"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/tracker"
	"github.com/cryptopunkscc/astrald/nodeinfo"
	"io"
	"time"
)

type cmdFunc func(io.ReadWriter, node.Node, []string) error
type cmdMap map[string]cmdFunc

var commands cmdMap

func add(_ io.ReadWriter, node node.Node, args []string) error {
	if len(args) < 1 {
		return errors.New("argument missing")
	}

	nodeInfo, err := nodeinfo.Parse(args[0])
	if err != nil {
		return err
	}

	if nodeInfo.Identity.IsEqual(node.Identity()) {
		return errors.New("cannot add self")
	}

	nodeinfo.AddToTracker(nodeInfo, node.Tracker())
	return nodeinfo.AddToContacts(nodeInfo, node.Contacts())
}

func forget(w io.ReadWriter, node node.Node, args []string) error {
	if len(args) == 0 {
		return errors.New("missing node id")
	}

	identity, err := node.Contacts().ResolveIdentity(args[0])
	if err != nil {
		return err
	}

	if err := node.Tracker().ForgetIdentity(identity); err != nil {
		return err
	}

	return node.Contacts().Delete(identity)
}

func info(w io.ReadWriter, node node.Node, _ []string) error {
	nodeInfo := nodeinfo.FromNode(node)

	fmt.Fprintln(w, "nodeID   ", node.Identity())
	fmt.Fprintln(w, "alias    ", node.Alias())
	fmt.Fprintln(w, "nodeinfo ", nodeInfo)
	for _, e := range node.Infra().Endpoints() {
		e, _ := node.Infra().Unpack(e.Network(), e.Pack())
		printEndpoint(w, e)
	}
	return nil
}

func cmdContacts(w io.ReadWriter, node node.Node, _ []string) error {
	for c := range node.Contacts().All() {
		fmt.Fprintln(w, "node", c.DisplayName())
		fmt.Fprintln(w, "pubkey", c.Identity().PublicKeyHex())

		endpoints, err := node.Tracker().AddrByIdentity(c.Identity())
		if err != nil {
			return err
		}

		for _, e := range endpoints {
			e, err := node.Infra().Unpack(e.Network(), e.Pack())
			if err != nil {
				continue
			}
			printEndpoint(w, e)
		}
		fmt.Fprintln(w)
	}

	return nil
}

func parse(w io.ReadWriter, node node.Node, args []string) error {
	if len(args) < 1 {
		return errors.New("argument missing")
	}

	info, err := nodeinfo.Parse(args[0])
	if err != nil {
		return err
	}

	fmt.Fprintf(w, "node %s\n", info.Identity.String())
	if info.Alias != "" {
		fmt.Fprintf(w, "  alias     %s\n", info.Alias)
	}
	for _, a := range info.Endpoints {
		addr, err := node.Infra().Unpack(a.Network(), a.Pack())
		if err != nil {
			continue
		}
		printEndpoint(w, addr)
	}

	return nil
}

func printEndpoint(w io.Writer, endpoint net.Endpoint) (int, error) {
	switch e := endpoint.(type) {
	case tracker.Addr:
		d := e.ExpiresAt.Sub(time.Now()).Round(time.Second)
		return fmt.Fprintf(w, "  %-10s%s (expires in %s)\n", e.Network(), e.String(), d)
	}
	return fmt.Fprintf(w, "  %-10s%s\n", endpoint.Network(), endpoint.String())
}

func init() {
	commands = cmdMap{
		"contacts": cmdContacts,
		"info":     info,
		"parse":    parse,
		"add":      add,
		"forget":   forget,
	}
}
