package admin

import (
	"context"
	"errors"
	"flag"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node/link"
)

func (cmd *CmdNet) link(term *Terminal, args []string) error {
	flags := flag.NewFlagSet("net link <nodeID>", flag.ContinueOnError)
	flags.SetOutput(term)
	flags.Usage = func() {
		term.Printf("Usage:\n\n  net link [options] <nodeID>\n\nOptions:\n")
		flags.PrintDefaults()
	}
	var network = flags.String("n", "", "link via this network only")
	var timeout = flags.Duration("t", defaultLinkTimeout, "set timeout")
	var addr = flags.String("a", "", "link via this address (requires -n)")
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

	var endpoints []net.Endpoint

	remoteID, err := cmd.mod.node.Resolver().Resolve(args[0])
	if err != nil {
		return err
	}

	if *addr != "" {
		if *network == "" {
			return errors.New("linking via address requires specifying the network")
		}
		e, err := cmd.mod.node.Infra().Parse(*network, *addr)
		if err != nil {
			return err
		}
		endpoints = []net.Endpoint{e}
	} else {
		endpoints, err = cmd.mod.node.Tracker().EndpointsByIdentity(remoteID)
		if err != nil {
			return err
		}

		if *network != "" {
			endpoints = selectEndpoints(endpoints, func(e net.Endpoint) bool {
				return e.Network() == *network
			})
		}
	}

	if len(endpoints) == 0 {
		return errors.New("no usable endpoints")
	}

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	lnk, err := link.MakeLink(ctx, cmd.mod.node, remoteID, link.Opts{Endpoints: endpoints})
	if err != nil {
		return err
	}

	err = cmd.mod.node.Network().AddLink(lnk)
	if err != nil {
		lnk.Close()
		return err
	}

	term.Printf("linked via %s\n", net.Network(lnk))

	return nil
}
