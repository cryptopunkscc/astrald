package admin

import (
	"fmt"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/tracker"
	"io"
	"time"
)

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

func printEndpoint(w io.Writer, endpoint net.Endpoint) (int, error) {
	switch e := endpoint.(type) {
	case tracker.Addr:
		d := e.ExpiresAt.Sub(time.Now()).Round(time.Second)
		return fmt.Fprintf(w, "  %-10s%s (expires in %s)\n", e.Network(), e.String(), d)
	}
	return fmt.Fprintf(w, "  %-10s%s\n", endpoint.Network(), endpoint.String())
}
