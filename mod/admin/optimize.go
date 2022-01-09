package admin

import (
	"errors"
	"github.com/cryptopunkscc/astrald/node"
	"io"
	"time"
)

func optimize(out io.ReadWriter, node *node.Node, args []string) error {
	if len(args) < 1 {
		return errors.New("missing arguments")
	}

	remoteID, err := node.Contacts.ResolveIdentity(args[0])
	if err != nil {
		return err
	}

	d := 30 * time.Second

	if len(args) > 1 {
		d, err = time.ParseDuration(args[1])
		if err != nil {
			return err
		}
	}

	node.Linking.Optimize(remoteID, d)

	return nil
}
