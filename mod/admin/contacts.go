package admin

import (
	"fmt"
	"github.com/cryptopunkscc/astrald/infra"
	"github.com/cryptopunkscc/astrald/node"
	"io"
)

func cmdContacts(w io.ReadWriter, node *node.Node, _ []string) error {
	for c := range node.Contacts.All() {
		fmt.Fprintln(w, "node", c.DisplayName())
		fmt.Fprintln(w, "pubkey", c.Identity().PublicKeyHex())

		for addr := range c.Addr(nil) {
			printAddr(w, addr)
		}
		fmt.Fprintln(w)
	}

	return nil
}

func printAddr(w io.Writer, addr infra.Addr) (int, error) {
	return fmt.Fprintf(w, "  %-10s%s\n", addr.Network(), addr.String())
}
