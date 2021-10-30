package admin

import (
	"context"
	"errors"
	"fmt"
	_f "github.com/cryptopunkscc/astrald/logfmt"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/route"
	"io"
	"time"
)

func peers(stream io.ReadWriter, node *node.Node, _ []string) error {
	for peer := range node.Network.Peers() {
		fmt.Fprintf(stream, "peer %s (%s in, %s out, %s idle)\n",
			_f.ID(peer.ID()),
			_f.DataSize(peer.BytesRead()),
			_f.DataSize(peer.BytesWritten()),
			peer.Idle().Round(time.Second),
		)
		for link := range peer.Links() {
			fmt.Fprintf(stream,
				"  %3s %s %s (%s in, %s out, %s idle)\n",
				_f.Dir(link.Outbound()),
				link.RemoteAddr().Network(),
				link.RemoteAddr().String(),
				_f.DataSize(link.BytesRead()),
				_f.DataSize(link.BytesWritten()),
				link.Idle().Round(time.Second),
			)
			for c := range link.Conns() {
				fmt.Fprintf(stream,
					"    %s: %s (%s in, %s out, %s idle)\n",
					_f.Dir(c.Outbound()),
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

func link(stream io.ReadWriter, node *node.Node, args []string) error {
	if len(args) < 1 {
		return errors.New("argument missing")
	}

	remoteID, err := node.ResolveIdentity(args[0])
	if err != nil {
		return err
	}

	link, err := node.Link(context.Background(), remoteID)
	if err != nil {
		return err
	}

	fmt.Fprintln(stream, "linked via", link.RemoteAddr().String())

	return nil
}

func routes(stream io.ReadWriter, node *node.Node, _ []string) error {
	for r := range node.Router.Routes() {
		writeRoute(stream, r)
		fmt.Fprintln(stream, "---")
	}
	return nil
}

func info(stream io.ReadWriter, node *node.Node, _ []string) error {
	writeRoute(stream, node.Route(false))
	fmt.Fprintln(stream, "pubroute ", node.Route(true))
	fmt.Fprintln(stream, "route    ", node.Route(false))
	return nil
}

func parse(stream io.ReadWriter, _ *node.Node, args []string) error {
	if len(args) < 1 {
		return errors.New("argument missing")
	}

	r, err := route.Parse(args[0])
	if err != nil {
		return err
	}

	fmt.Fprintf(stream, "%-10s%s\n", "node", r.Identity.String())
	for _, addr := range r.Addresses {
		fmt.Fprintf(stream, "%-10s%s\n", addr.Network(), addr.String())
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

	node.Router.AddRoute(r)

	return nil
}

func writeRoute(w io.Writer, r *route.Route) {
	fmt.Fprintf(w, "%-10s%s\n", "node", r.Identity.String())
	for _, addr := range r.Addresses {
		fmt.Fprintf(w, "%-10s%s\n", addr.Network(), addr.String())
	}
}
