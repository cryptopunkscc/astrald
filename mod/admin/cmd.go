package admin

import (
	"context"
	"errors"
	"fmt"
	"github.com/cryptopunkscc/astrald/infra"
	_f "github.com/cryptopunkscc/astrald/logfmt"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/network/contacts"
	"io"
	"time"
)

func peers(w io.ReadWriter, node *node.Node, _ []string) error {
	for peer := range node.Network.Peers() {
		peerID := _f.ID(peer.Identity())
		if a := node.Network.Contacts.GetAlias(peer.Identity()); a != "" {
			peerID = a
		}

		fmt.Fprintf(w, "peer %s (idle %s)\n",
			peerID,
			peer.Idle().Round(time.Second),
		)
		for link := range peer.Links() {
			fmt.Fprintf(w,
				"  %s %s %s (idle %s, lat %.1fms)\n",
				_f.Bool(link.Outbound(), "=>", "<="),
				link.RemoteAddr().Network(),
				link.RemoteAddr().String(),
				link.Idle().Round(time.Second),
				float64(link.Latency().Microseconds())/1000,
			)
			for c := range link.Conns() {
				fmt.Fprintf(w,
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

	node.Network.Connect(context.Background(), node.Network.Peer(remoteID))
	if err != nil {
		return err
	}

	return nil
}

func graph(w io.ReadWriter, node *node.Node, _ []string) error {
	for nodeID := range node.Network.Contacts.Identities() {
		fmt.Fprintln(w, "node", nodeID.PublicKeyHex())
		fmt.Fprintln(w, "alias", node.Network.Contacts.GetAlias(nodeID))

		for addr := range node.Network.Contacts.Resolve(nodeID) {
			printAddr(w, addr)
		}
		fmt.Fprintln(w)
	}
	return nil
}

func info(w io.ReadWriter, node *node.Node, _ []string) error {
	fmt.Fprintln(w, "nodeID   ", node.Identity)
	fmt.Fprintln(w, "alias    ", node.Network.Alias())
	fmt.Fprintln(w, "pubinfo  ", node.Network.Info(true))
	fmt.Fprintln(w, "info     ", node.Network.Info(false))
	for _, addr := range node.Network.Info(false).Addresses {
		printAddr(w, addr)
	}
	return nil
}

func parse(w io.ReadWriter, _ *node.Node, args []string) error {
	if len(args) < 1 {
		return errors.New("argument missing")
	}

	info, err := contacts.ParseInfo(args[0])
	if err != nil {
		return err
	}

	fmt.Fprintln(w, "node", info.Identity.String())
	if info.Alias != "" {
		fmt.Fprintln(w, "alias", info.Alias)
	}
	for _, addr := range info.Addresses {
		printAddr(w, addr)
	}

	return nil
}

func add(_ io.ReadWriter, node *node.Node, args []string) error {
	if len(args) < 1 {
		return errors.New("argument missing")
	}

	r, err := contacts.ParseInfo(args[0])
	if err != nil {
		return err
	}

	node.Network.Contacts.AddInfo(r)

	return nil
}

func printAddr(w io.Writer, addr infra.Addr) (int, error) {
	return fmt.Fprintf(w, "  %-10s%s\n", addr.Network(), addr.String())
}
