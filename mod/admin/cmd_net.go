package admin

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/node/network"
	"time"
)

const defaultLinkTimeout = time.Minute

var _ Command = &CmdNet{}

type CmdNet struct {
	mod *Module
}

func (cmd *CmdNet) Exec(term *Terminal, args []string) error {
	if len(args) < 2 {
		return cmd.help(term)
	}

	switch args[1] {
	case "link":
		return cmd.link(term, args[2:])

	case "unlink":
		return cmd.unlink(term, args[2:])

	case "ls":
		return cmd.ls(term, args[2:])

	case "check":
		return cmd.check(term, args[2:])

	case "help":
		return cmd.help(term)

	default:
		return errors.New("invalid command")
	}
}

func (cmd *CmdNet) link(term *Terminal, args []string) error {
	if len(args) < 1 {
		return cmd.linkHelp(term)
	}

	remoteID, err := cmd.mod.node.Resolver().Resolve(args[0])
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

	term.Printf("linked via %s\n", link.Network())

	return nil
}

func (cmd *CmdNet) linkHelp(term *Terminal) error {
	term.Printf("help: net link <node>\n\n")
	return nil
}

func (cmd *CmdNet) unlink(term *Terminal, args []string) error {
	if len(args) < 1 {
		return cmd.unlinkHelp(term)
	}

	identity, err := cmd.mod.node.Resolver().Resolve(args[0])
	if err != nil {
		return err
	}

	var peer = cmd.mod.node.Network().Peers().Find(identity)
	if peer == nil {
		return errors.New("peer not linked")
	}

	peer.Unlink()

	term.Printf("unlinked\n")

	return nil
}

func (cmd *CmdNet) unlinkHelp(term *Terminal) error {
	term.Printf("help: net unlink <node>\n\n")
	return nil
}

func (cmd *CmdNet) ls(term *Terminal, _ []string) error {
	for _, peer := range cmd.mod.node.Network().Peers().All() {
		term.Printf("peer %s (idle %s)\n",
			peer.Identity(),
			peer.Idle().Round(time.Second),
		)
		for _, link := range peer.Links() {
			term.Printf("  %s %s %s (idle %s, age %s, prio %d, ping %f)\n",
				DoubleArrow(link.Outbound()),
				link.RemoteEndpoint().Network(),
				link.RemoteEndpoint(),
				link.Activity().Idle().Round(time.Second),
				time.Since(link.EstablishedAt()).Round(time.Second),
				link.Priority(),
				link.Health().AverageRTT().Round(100*time.Microsecond),
			)
			for _, c := range link.Conns().All() {
				term.Printf("    %s %s [%d:%d] (idle %s)\n",
					Arrow(c.Outbound()),
					c.Query(),
					c.LocalPort(),
					c.RemotePort(),
					c.Idle().Round(time.Second),
				)
			}
		}

		term.Printf("\n")
	}

	if n, ok := cmd.mod.node.Network().(*network.CoreNetwork); ok {
		for _, l := range n.Linkers() {
			peerName := cmd.mod.node.Resolver().DisplayName(l)
			term.Printf("peer %s linking...\n", peerName)
		}
	}

	return nil
}

func (cmd *CmdNet) check(_ *Terminal, _ []string) error {
	for _, peer := range cmd.mod.node.Network().Peers().All() {
		peer.Check()
	}
	return nil
}

func (cmd *CmdNet) help(term *Terminal) error {
	term.Printf("help: net <command> [options]\n\n")
	term.Printf("commands:\n\n")
	term.Printf("  ls        list linked nodes\n")
	term.Printf("  link      link a node\n")
	term.Printf("  unlink    unlink a node\n")
	term.Printf("  check     run health check on all links\n")
	term.Printf("  help      show help\n")
	return nil
}

func (cmd *CmdNet) ShortDescription() string {
	return "manage p2p network"
}
