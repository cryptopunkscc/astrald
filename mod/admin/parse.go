package admin

import (
	"errors"
	"fmt"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/contacts"
	"io"
)

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
