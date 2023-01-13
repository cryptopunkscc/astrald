package admin

import (
	"fmt"
	"github.com/cryptopunkscc/astrald/infra"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/tracker"
	"io"
	"time"
)

func cmdContacts(w io.ReadWriter, node *node.Node, _ []string) error {
	for c := range node.Contacts.All() {
		fmt.Fprintln(w, "node", c.DisplayName())
		fmt.Fprintln(w, "pubkey", c.Identity().PublicKeyHex())

		addrs, err := node.Tracker.AddrByIdentity(c.Identity())
		if err != nil {
			return err
		}

		for _, a := range addrs {
			addr, err := node.Infra.Unpack(a.Network(), a.Pack())
			if err != nil {
				continue
			}
			printAddr(w, addr)
		}
		fmt.Fprintln(w)
	}

	return nil
}

func printAddr(w io.Writer, addr infra.Addr) (int, error) {
	switch addr := addr.(type) {
	case tracker.Addr:
		d := addr.ExpiresAt.Sub(time.Now()).Round(time.Second)
		return fmt.Fprintf(w, "  %-10s%s (expires in %s)\n", addr.Network(), addr.String(), d)
	}
	return fmt.Fprintf(w, "  %-10s%s\n", addr.Network(), addr.String())
}
