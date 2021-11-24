package admin

import (
	"errors"
	"fmt"
	"github.com/cryptopunkscc/astrald/infra"
	_f "github.com/cryptopunkscc/astrald/logfmt"
	"github.com/cryptopunkscc/astrald/node"
	_graph "github.com/cryptopunkscc/astrald/node/network/graph"
	"io"
	"time"
)

func peers(ui io.ReadWriter, node *node.Node, _ []string) error {
	for peer := range node.Network.Each() {
		fmt.Fprintf(ui, "peer %s (idle %s)\n",
			_f.ID(peer.Identity()),
			peer.Idle().Round(time.Second),
		)
		for link := range peer.Links.Links() {
			fmt.Fprintf(ui,
				"  %s %s %s (idle %s)\n",
				_f.Bool(link.Outbound(), "=>", "<="),
				link.RemoteAddr().Network(),
				link.RemoteAddr().String(),
				link.Idle().Round(time.Second),
			)
			for c := range link.Conns() {
				fmt.Fprintf(ui,
					"    %s %s (idle %s)\n",
					_f.Bool(c.Outbound(), "->", "<-"),
					c.Query(),
					c.Idle().Round(time.Second),
				)
			}
		}
	}

	return nil
}

func link(_ io.ReadWriter, node *node.Node, args []string) error {
	if len(args) < 1 {
		return errors.New("argument missing")
	}

	remoteID, err := node.ResolveIdentity(args[0])
	if err != nil {
		return err
	}

	node.Network.Linker.Wake(remoteID)
	if err != nil {
		return err
	}

	return nil
}

func graph(w io.ReadWriter, node *node.Node, _ []string) error {
	for nodeID := range node.Network.Graph.Nodes() {
		fmt.Fprintln(w, "node", nodeID.PublicKeyHex())
		for addr := range node.Network.Graph.Resolve(nodeID) {
			printAddr(w, addr)
		}
		fmt.Fprintln(w)
	}
	return nil
}

func info(w io.ReadWriter, node *node.Node, _ []string) error {
	fmt.Fprintln(w, "nodeID   ", node.Identity)
	fmt.Fprintln(w, "pubinfo  ", node.Network.Info(true))
	fmt.Fprintln(w, "info     ", node.Network.Info(false))
	for _, addr := range node.Network.Info(false).Addresses {
		printAddr(w, addr)
	}
	return nil
}

func parse(ui io.ReadWriter, _ *node.Node, args []string) error {
	if len(args) < 1 {
		return errors.New("argument missing")
	}

	r, err := _graph.Parse(args[0])
	if err != nil {
		return err
	}

	fmt.Fprintf(ui, "%-10s%s\n", "node", r.Identity.String())
	for _, addr := range r.Addresses {
		fmt.Fprintf(ui, "%-10s%s\n", addr.Network(), addr.String())
	}

	return nil
}

func add(_ io.ReadWriter, node *node.Node, args []string) error {
	if len(args) < 1 {
		return errors.New("argument missing")
	}

	r, err := _graph.Parse(args[0])
	if err != nil {
		return err
	}

	node.Network.Graph.AddInfo(r)

	return nil
}

func printAddr(w io.Writer, addr infra.Addr) (int, error) {
	return fmt.Fprintf(w, "  %-10s%s\n", addr.Network(), addr.String())
}
