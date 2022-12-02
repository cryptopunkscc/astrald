package admin

import (
	"context"
	"errors"
	"fmt"
	"github.com/cryptopunkscc/astrald/node"
	"io"
	"time"
)

func link(out io.ReadWriter, node *node.Node, args []string) error {
	if len(args) < 1 {
		return errors.New("missing arguments")
	}

	remoteID, err := node.Contacts.ResolveIdentity(args[0])
	if err != nil {
		return err
	}

	d := 10 * time.Second

	if len(args) > 1 {
		d, err = time.ParseDuration(args[1])
		if err != nil {
			return err
		}
	}

	timeout, _ := context.WithTimeout(context.Background(), d)

	_, err = node.Peers.Link(timeout, remoteID)
	if err != nil {
		return err
	}

	fmt.Fprintf(out, "linked\n")

	return nil
}
