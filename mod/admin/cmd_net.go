package admin

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/link"
	"github.com/cryptopunkscc/astrald/sig"
	"io"
	"reflect"
	"strconv"
	"strings"
	"time"
)

const defaultLinkTimeout = time.Minute

var _ Command = &CmdNet{}

type CmdNet struct {
	mod *Module
}

type checkLatency interface {
	Latency() time.Duration
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

	case "show":
		return cmd.show(term, args[2:])

	case "close":
		return cmd.close(term, args[2:])

	case "conns":
		return cmd.conns(term, args[2:])

	case "conn":
		return cmd.conn(term, args[2:])

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

	var f1 = "%-6d %-30s %-20s %-8s %8s %8s\n"
	var f2 = "%-6d %s %-20s %-8s %8s %8s\n"

	term.Printf(f1, Header("ID"), Header("Identities"), Header("Query"), Header("State"), Header("In"), Header("Out"))
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
			log.DataSize(conn.BytesIn()).HumanReadable(),
			log.DataSize(conn.BytesOut()).HumanReadable(),
		)
	}

	return nil
}

func (cmd *CmdNet) conn(term *Terminal, args []string) error {
	corenode, ok := cmd.mod.node.(*node.CoreNode)
	if !ok {
		return errors.New("unsupported node type")
	}

	if len(args) < 1 {
		term.Printf("usage: net conn <ConnID>\n")
		return nil
	}

	cid, err := strconv.Atoi(args[0])
	if err != nil {
		return err
	}

	conn := corenode.Conns().Find(cid)
	if conn == nil {
		return errors.New("no such connection")
	}

	term.Printf("Caller:\n")
	cmd.printWriterInfo(term, conn.Caller())

	term.Printf("Target:\n")
	cmd.printWriterInfo(term, conn.Target())

	return nil
}

func (cmd *CmdNet) printWriterInfo(term *Terminal, writer io.Writer) {
	final := net.FinalWriter(writer)
	term.Printf("Final: %s\n", reflect.TypeOf(final))

	if w, ok := final.(*link.PortWriter); ok {
		term.Printf("Port: %d\n", w.Port())
	}

	if w, ok := final.(*net.PipeWriter); ok {
		term.Printf("Source: %d\n", reflect.TypeOf(w.Source()))
		if b, ok := w.Source().(*link.PortBinding); ok {
			term.Printf("Port: %d\n", b.Port())

		}
	}
}

func (cmd *CmdNet) show(term *Terminal, args []string) error {
	if len(args) < 1 {
		term.Printf("usage: net show <linkID>")
		return nil
	}

	lid, err := strconv.Atoi(args[0])
	if err != nil {
		return err
	}

	l, err := cmd.mod.node.Network().Links().Find(lid)
	if err != nil {
		return err
	}

	term.Printf("ID:               %v (%v)\n", l.ID(), getLinkType(l.Link))
	term.Printf("Identities:       %v:%v\n", l.LocalIdentity(), l.RemoteIdentity())
	if t := l.Transport(); t != nil {
		term.Printf("Network:          %v\n", net.Network(l))
		term.Printf("Local endpoint:   %v\n", t.LocalEndpoint())
		term.Printf("Remote endpoint:  %v\n", t.RemoteEndpoint())
		term.Printf("Outbound:         %v\n", t.Outbound())
	}
	if l, ok := l.Link.(checkLatency); ok {
		term.Printf("Latency:          %v\n", l.Latency().Round(time.Millisecond))
	}
	term.Printf("Age:              %v (%v)\n",
		time.Since(l.AddedAt()).Round(time.Second),
		l.AddedAt(),
	)
	if idler, ok := l.Link.(sig.Idler); ok {
		term.Printf("Idle:             %v\n", idler.Idle().Round(time.Second))
	}
	return nil
}

func (cmd *CmdNet) close(term *Terminal, args []string) error {
	if len(args) < 1 {
		term.Printf("usage: net close <linkID>")
		return nil
	}

	lid, err := strconv.Atoi(args[0])
	if err != nil {
		return err
	}

	l, err := cmd.mod.node.Network().Links().Find(lid)
	if err != nil {
		return err
	}

	return l.Close()
}

func (cmd *CmdNet) links(term *Terminal, _ []string) error {
	var f = "%-8d %-24s %-24s %-8s %10s %10s %10s\n"

	term.Printf(f, Header("ID"), Header("Local"), Header("Remote"), Header("Net"), Header("Idle"), Header("Age"), Header("Ping"))
	for _, l := range cmd.mod.node.Network().Links().All() {
		if l == nil {
			term.Printf("[nil link]\n")
			continue
		}
		var idle time.Duration = -1
		var lat time.Duration = -1

		if i, ok := l.Link.(sig.Idler); ok {
			idle = i.Idle().Round(time.Second)
		}

		if l, ok := l.Link.(checkLatency); ok {
			lat = l.Latency()
		}

		term.Printf(f,
			l.ID(),
			l.LocalIdentity(),
			l.RemoteIdentity(),
			Keyword(net.Network(l)),
			idle,
			time.Since(l.AddedAt()).Round(time.Second),
			lat.Round(time.Millisecond),
		)
	}

	return nil
}

func (cmd *CmdNet) check(_ *Terminal, _ []string) error {
	type checker interface {
		Check()
	}

	for _, l := range cmd.mod.node.Network().Links().All() {
		if c, ok := l.Link.(checker); ok {
			c.Check()
		}
	}
	return nil
}

func (cmd *CmdNet) help(term *Terminal) error {
	term.Printf("help: net <command> [options]\n\n")
	term.Printf("commands:\n\n")
	term.Printf("  links     list all links\n")
	term.Printf("  show      show link info\n")
	term.Printf("  close     close a link\n")
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

func getLinkType(l any) string {
	var t = reflect.TypeOf(l)
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.Name()
}
