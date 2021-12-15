package admin

import (
	"errors"
	"fmt"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/nodeinfo"
	"io"
)

func parse(w io.ReadWriter, _ *node.Node, args []string) error {
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
	for _, addr := range info.Addresses {
		printAddr(w, addr)
	}

	return nil
}
