package admin

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/node/infra/drivers/gw"
	"github.com/cryptopunkscc/astrald/node/network"
	"time"
)

const defaultLinkTimeout = time.Minute

var _ Command = &NetCommand{}

type NetCommand struct {
	mod *Module
}

func (cmd *NetCommand) Exec(t *Terminal, args []string) error {
	if len(args) < 2 {
		return cmd.usage(t)
	}

	switch args[1] {
	case "link":
		return cmd.link(t, args[2:])

	case "unlink":
		return cmd.unlink(t, args[2:])

	case "peers":
		return cmd.peers(t, args[2:])

	case "check":
		return cmd.check(t, args[2:])

	default:
		return errors.New("invalid command")
	}
}

func (cmd *NetCommand) link(t *Terminal, args []string) error {
	if len(args) < 1 {
		return cmd.linkUsage(t)
	}

	remoteID, err := cmd.mod.node.Contacts().ResolveIdentity(args[0])
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

	link, err := cmd.mod.node.Network().Link(ctx, remoteID)
	if err != nil {
		return err
	}

	t.Printf("linked via %s\n", link.Network())

	return nil
}

func (cmd *NetCommand) linkUsage(t *Terminal) error {
	t.Printf("usage: net link <node>\n\n")
	return nil
}

func (cmd *NetCommand) unlink(t *Terminal, args []string) error {
	if len(args) < 1 {
		return cmd.unlinkUsage(t)
	}

	identity, err := cmd.mod.node.Contacts().ResolveIdentity(args[0])
	if err != nil {
		return err
	}

	peer := cmd.mod.node.Network().Peers().Find(identity)
	if peer == nil {
		return errors.New("peer not found")
	}

	peer.Unlink()

	t.Printf("unlinked\n")

	return nil
}

func (cmd *NetCommand) unlinkUsage(t *Terminal) error {
	t.Printf("usage: net unlink <node>\n\n")
	return nil
}

func (cmd *NetCommand) peers(t *Terminal, _ []string) error {
	for _, peer := range cmd.mod.node.Network().Peers().All() {
		peerName := cmd.mod.node.Contacts().DisplayName(peer.Identity())

		t.Printf("peer %s (idle %s)\n",
			peerName,
			peer.Idle().Round(time.Second),
		)
		for _, link := range peer.Links() {
			remoteAddr := link.RemoteEndpoint().String()

			if gwEndpoint, ok := link.RemoteEndpoint().(gw.Endpoint); ok {
				remoteAddr = "via " + cmd.mod.node.Contacts().DisplayName(gwEndpoint.Gate())
			}

			t.Printf("  %s %s %s (idle %s, age %s, prio %d, ping %.1fms)\n",
				log.Bool(link.Outbound(), "=>", "<="),
				link.RemoteEndpoint().Network(),
				remoteAddr,
				link.Activity().Idle().Round(time.Second),
				time.Since(link.EstablishedAt()).Round(time.Second),
				link.Priority(),
				float64(link.Health().AverageRTT().Microseconds())/1000,
			)
			for _, c := range link.Conns().All() {
				t.Printf("    %s %s [%d:%d] (idle %s)\n",
					log.Bool(c.Outbound(), "->", "<-"),
					c.Query(),
					c.LocalPort(),
					c.RemotePort(),
					c.Idle().Round(time.Second),
				)
			}
		}

		t.Printf("\n")
	}

	if n, ok := cmd.mod.node.Network().(*network.CoreNetwork); ok {
		for _, l := range n.Linkers() {
			peerName := cmd.mod.node.Contacts().DisplayName(l)
			t.Printf("peer %s linking...\n", peerName)
		}
	}

	return nil
}

func (cmd *NetCommand) check(t *Terminal, _ []string) error {
	for _, peer := range cmd.mod.node.Network().Peers().All() {
		peer.Check()
	}
	return nil
}

func (cmd *NetCommand) usage(t *Terminal) error {
	t.Printf("usage: net <command> [options]\n\n")
	t.Printf("commands:\n\n")
	t.Printf("  peers     list linked nodes\n")
	t.Printf("  link      link a node\n")
	t.Printf("  unlink    unlink a node\n")
	t.Printf("  check     run health check on all links\n")
	return nil
}

func (cmd *NetCommand) HelpDescription() string {
	return "manage p2p network"
}
