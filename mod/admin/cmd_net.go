package admin

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node"
	"reflect"
	"strings"
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

	case "links":
		return cmd.links(term, args[2:])

	case "conns":
		return cmd.conns(term, args[2:])

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

	term.Printf("linked via %s\n", net.Network(link))

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

	links := cmd.mod.node.Network().Links().ByRemoteIdentity(identity).All()
	if len(links) == 0 {
		return errors.New("peer not linked")
	}

	for _, l := range links {
		l.Close()
	}

	term.Printf("unlinked\n")

	return nil
}

func (cmd *CmdNet) unlinkHelp(term *Terminal) error {
	term.Printf("help: net unlink <node>\n\n")
	return nil
}

func (cmd *CmdNet) conns(term *Terminal, _ []string) error {
	corenode, ok := cmd.mod.node.(*node.CoreNode)
	if !ok {
		return errors.New("unsupported node type")
	}

	var f1 = "%-6d %-30s %-20s %-8s %-8s %8s %8s\n"
	var f2 = "%-6d %s %-20s %-8s %-8s %8s %8s\n"

	term.Printf(f1, Header("ID"), Header("Identities"), Header("Query"), Header("State"), Header("Origin"), Header("In"), Header("Out"))
	for _, conn := range corenode.Conns().All() {
		c := term.Color
		term.Color = false
		var peersWidth = len(term.Sprintf(term.Sprintf("%s:%s", conn.Caller().Identity(), conn.Target().Identity())))
		term.Color = c
		var peers = term.Sprintf("%s:%s", conn.Caller().Identity(), conn.Target().Identity())
		if peersWidth < 30 {
			peers = peers + strings.Repeat(" ", 30-peersWidth)
		}

		term.Printf(f2,
			conn.ID(),
			peers,
			conn.Query().Query(),
			conn.State(),
			conn.Query().Origin(),
			log.DataSize(conn.BytesIn()).HumanReadable(),
			log.DataSize(conn.BytesOut()).HumanReadable(),
		)
	}

	return nil
}

func (cmd *CmdNet) links(term *Terminal, _ []string) error {
	var f = "%-24s %-24s %-8s %-16s\n"

	term.Printf(f, Header("Local"), Header("Remote"), Header("Net"), Header("Type"))
	for _, l := range cmd.mod.node.Network().Links().All() {
		term.Printf(f,
			l.LocalIdentity(),
			l.RemoteIdentity(),
			Keyword(net.Network(l)),
			Keyword(reflect.TypeOf(l).String()),
		)
	}

	return nil
}

func (cmd *CmdNet) check(_ *Terminal, _ []string) error {
	type checker interface {
		Check()
	}

	for _, l := range cmd.mod.node.Network().Links().All() {
		if c, ok := l.(checker); ok {
			c.Check()
		}
	}
	return nil
}

func (cmd *CmdNet) help(term *Terminal) error {
	term.Printf("help: net <command> [options]\n\n")
	term.Printf("commands:\n\n")
	term.Printf("  links     list all links\n")
	term.Printf("  link      link a node\n")
	term.Printf("  unlink    unlink a node\n")
	term.Printf("  conns     list all connections\n")
	term.Printf("  check     run health check on all links\n")
	term.Printf("  help      show help\n")
	return nil
}

func (cmd *CmdNet) ShortDescription() string {
	return "manage p2p network"
}
