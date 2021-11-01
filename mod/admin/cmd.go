package admin

import (
	"errors"
	"fmt"
	_f "github.com/cryptopunkscc/astrald/logfmt"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/network/route"
	"io"
	"time"
)

func peers(ui io.ReadWriter, node *node.Node, _ []string) error {
	for peer := range node.Network.All() {
		fmt.Fprintf(ui, "peer %s (%s in, %s out, last seen %s ago)\n",
			_f.ID(peer.Identity()),
			_f.DataSize(peer.BytesRead()),
			_f.DataSize(peer.BytesWritten()),
			peer.Idle().Round(time.Second),
		)
		for link := range peer.Links() {
			fmt.Fprintf(ui,
				"  %s %s %s (%s in, %s out, %s idle)\n",
				_f.Bool(link.Outbound(), "=>", "<="),
				link.RemoteAddr().Network(),
				link.RemoteAddr().String(),
				_f.DataSize(link.BytesRead()),
				_f.DataSize(link.BytesWritten()),
				link.Idle().Round(time.Second),
			)
			for c := range link.Conns() {
				fmt.Fprintf(ui,
					"    %s %s (%s in, %s out, %s idle)\n",
					_f.Bool(c.Outbound(), "->", "<-"),
					c.Query(),
					_f.DataSize(c.BytesRead()),
					_f.DataSize(c.BytesWritten()),
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

func routes(ui io.ReadWriter, node *node.Node, _ []string) error {
	for r := range node.Network.Router.Routes() {
		printRoute(ui, r)
		fmt.Fprintln(ui, "")
	}
	return nil
}

func info(ui io.ReadWriter, node *node.Node, _ []string) error {
	printRoute(ui, node.Network.Route(false))
	fmt.Fprintln(ui, "pubroute ", node.Network.Route(true))
	fmt.Fprintln(ui, "route    ", node.Network.Route(false))
	return nil
}

func parse(ui io.ReadWriter, _ *node.Node, args []string) error {
	if len(args) < 1 {
		return errors.New("argument missing")
	}

	r, err := route.Parse(args[0])
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

	r, err := route.Parse(args[0])
	if err != nil {
		return err
	}

	node.Network.Router.AddRoute(r)

	return nil
}

func printRoute(w io.Writer, r *route.Route) {

	fmt.Fprintf(w, "%s\n", r.String())
	fmt.Fprintf(w, "%-10s%s\n", "node", r.Identity.String())
	for _, addr := range r.Addresses {
		fmt.Fprintf(w, "%-10s%s\n", addr.Network(), addr.String())
	}
}
