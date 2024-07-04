package admin

import (
	"context"
	"errors"
	"flag"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/net"
	"time"
)

const defaultLinkTimeout = time.Minute

func (cmd *CmdNet) link(term admin.Terminal, args []string) error {
	flags := flag.NewFlagSet("net link <nodeID>", flag.ContinueOnError)
	flags.SetOutput(term)
	flags.Usage = func() {
		term.Printf("Usage:\n\n  net link [options] <nodeID>\n\nOptions:\n")
		flags.PrintDefaults()
	}
	var timeout = flags.Duration("t", defaultLinkTimeout, "set timeout")
	err := flags.Parse(args)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	args = flags.Args()

	if len(args) < 1 {
		flags.Usage()
		return nil
	}

	remoteID, err := cmd.mod.node.Resolver().Resolve(args[0])
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	lnk, err := cmd.mod.node.Network().Link(ctx, remoteID)

	if err != nil {
		return err
	}

	term.Printf("linked via %s\n", net.Network(lnk))

	return nil
}
