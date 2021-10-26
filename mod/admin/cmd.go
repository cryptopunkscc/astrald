package admin

import (
	"context"
	"errors"
	"fmt"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/logfmt"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/route"
	"io"
)

func links(stream io.ReadWriter, node *node.Node, _ []string) error {
	fmt.Fprintf(stream, "active links:\n")

	for link := range node.Network.View.Links() {
		fmt.Fprintf(stream,
			"%s %s via %s %s\n",
			link.RemoteIdentity().String(),
			logfmt.Dir(link.Outbound()),
			link.RemoteAddr().Network(),
			link.RemoteAddr().String(),
		)
		for c := range link.Connections() {
			fmt.Fprintf(stream,
				"  %s (%s, %d bytes in, %d bytes out)\n",
				c.Query(),
				logfmt.Dir(c.Outbound()),
				c.BytesRead(),
				c.BytesWritten())
		}
	}
	return nil
}

func link(stream io.ReadWriter, node *node.Node, args []string) error {
	if len(args) < 1 {
		return errors.New("argument missing")
	}

	remoteID, err := id.ParsePublicKeyHex(args[0])
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
	for _, r := range node.Routes {
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

	node.AddRoute(r)

	return nil
}

func writeRoute(w io.Writer, r *route.Route) {
	fmt.Fprintf(w, "%-10s%s\n", "node", r.Identity.String())
	for _, addr := range r.Addresses {
		fmt.Fprintf(w, "%-10s%s\n", addr.Network(), addr.String())
	}
}
