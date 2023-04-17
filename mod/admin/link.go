package admin

import (
	"context"
	"errors"
	"fmt"
	"github.com/cryptopunkscc/astrald/node"
	"io"
	"time"
)

const defaultLinkTimeout = time.Minute

func link(out io.ReadWriter, node node.Node, args []string) error {
	if len(args) < 1 {
		return errors.New("missing arguments")
	}

	remoteID, err := node.Contacts().ResolveIdentity(args[0])
	if err != nil {
		return err
	}

	timeout := defaultLinkTimeout

	if len(args) > 1 {
		timeout, err = time.ParseDuration(args[1])
		if err != nil {
			return err
		}
	}

	ctx, _ := context.WithTimeout(context.Background(), timeout)

	link, err := node.Network().Link(ctx, remoteID)
	if err != nil {
		return err
	}

	fmt.Fprintf(out, "linked via %s\n", link.Network())

	return nil
}
