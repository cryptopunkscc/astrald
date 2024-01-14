package admin

import (
	"cmp"
	"errors"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/link"
	"github.com/cryptopunkscc/astrald/node/network"
	"github.com/cryptopunkscc/astrald/node/router"
	"github.com/cryptopunkscc/astrald/sig"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"time"
)

const defaultLinkTimeout = time.Minute

var _ admin.Command = &CmdNet{}

type CmdNet struct {
	mod *Module
}

type checkLatency interface {
	Latency() time.Duration
}

func (cmd *CmdNet) Exec(term admin.Terminal, args []string) error {
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

	case "routes":
		return cmd.routes(term, args[2:])

	case "conn":
		return cmd.conn(term, args[2:])

	case "check":
		return cmd.check(term, args[2:])

	case "reroute":
		return cmd.reroute(term, args[2:])

	case "help":
		return cmd.help(term)

	default:
		return errors.New("invalid command")
	}
}

func (cmd *CmdNet) unlink(term admin.Terminal, args []string) error {
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

func (cmd *CmdNet) unlinkHelp(term admin.Terminal) error {
	term.Printf("help: net unlink <node>\n\n")
	return nil
}

func (cmd *CmdNet) conns(term admin.Terminal, _ []string) error {
	corenode, ok := cmd.mod.node.(*node.CoreNode)
	if !ok {
		return errors.New("unsupported node type")
	}

	var f1 = "%-6d %-30s %-20s %-10s %8s %8s %-16s\n"
	var f2 = "%-6d %s %-20s %-10s %8s %8s %-16s\n"

	term.Printf(
		f1,
		admin.Header("ID"),
		admin.Header("Identities"),
		admin.Header("Query"),
		admin.Header("State"),
		admin.Header("In"),
		admin.Header("Out"),
		admin.Header("Nonce"),
	)
	for _, conn := range corenode.Conns().All() {
		c := term.Color()
		term.SetColor(false)
		var peersWidth = len(term.Sprintf(term.Sprintf("%s:%s", conn.Query().Caller(), conn.Query().Target())))
		term.SetColor(c)
		var peers = term.Sprintf("%s:%s", conn.Query().Caller(), conn.Query().Target())
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
			conn.Query().Nonce(),
		)
	}

	return nil
}

func (cmd *CmdNet) conn(term admin.Terminal, args []string) error {
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

	term.Printf("CALLER CHAIN\n")
	if conn.Caller() == nil {
		term.Printf("caller is nil\n")
	} else {
		cmd.printChainInfo(term, conn.Caller())
	}

	term.Printf("\nTARGET CHAIN\n")
	if conn.Target() == nil {
		term.Printf("target is nil\n")
	} else {
		cmd.printChainInfo(term, conn.Target())
	}
	return nil
}

func (cmd *CmdNet) printChainInfo(term admin.Terminal, element any) {
	i := net.RootSource(element)

	for i != nil {
		term.Printf("\n%s\n", admin.Keyword(reflect.TypeOf(i).Elem().Name()))

		switch w := i.(type) {
		case *link.PortBinding:
			term.Printf("  Identity: %d\n", w.Output().Identity())
			term.Printf("  Port: %d\n", w.Port())
			if t := w.Transport(); t != nil {
				var (
					network                       = "unknown"
					localEndpoint, remoteEndpoint string
				)
				if e := t.LocalEndpoint(); e != nil {
					network = e.Network()
					localEndpoint = e.String()
				}

				if e := t.RemoteEndpoint(); e != nil {
					remoteEndpoint = e.String()
				}

				term.Printf("  Transport: %s %s~%s\n",
					network,
					localEndpoint,
					remoteEndpoint,
				)
			}
			term.Printf("  Buffer: %d/%d\n", w.Used(), w.BufferSize())

		case *link.PortWriter:
			term.Printf("  Identity: %d\n", w.Identity())
			term.Printf("  Port: %d\n", w.Port())
			term.Printf("  Source: %s\n", reflect.TypeOf(w.Source()))
			if t := w.Transport(); t != nil {
				var (
					network                       = "unknown"
					localEndpoint, remoteEndpoint string
				)
				if e := t.LocalEndpoint(); e != nil {
					network = e.Network()
					localEndpoint = e.String()
				}

				if e := t.RemoteEndpoint(); e != nil {
					remoteEndpoint = e.String()
				}

				term.Printf("  Transport: %s %s~%s\n",
					network,
					localEndpoint,
					remoteEndpoint,
				)
			}
			term.Printf("  Buffer: %d\n", w.BufferSize())

		case *router.MonitoredWriter:
			term.Printf("  Identity: %d\n", w.Identity())
			term.Printf("  Bytes: %d\n", w.Bytes())

		case *net.SecurePipeWriter:
			term.Printf("  Identity: %d\n", w.Identity())
			term.Printf("  Transport: %s\n", reflect.TypeOf(w.Insecure()))

		case *net.IdentityTranslation:
			term.Printf("  Identity: %d\n", w.Identity())

		}

		if nw, ok := i.(net.OutputGetter); ok {
			i = nw.Output()
		} else {
			break
		}
	}
}

func (cmd *CmdNet) show(term admin.Terminal, args []string) error {
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
	term.Printf("Local identity:   %v (%v)\n", l.LocalIdentity(), admin.Faded(l.LocalIdentity().PublicKeyHex()))
	term.Printf("Remote identity:  %v (%v)\n", l.RemoteIdentity(), admin.Faded(l.RemoteIdentity().PublicKeyHex()))
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

func (cmd *CmdNet) close(term admin.Terminal, args []string) error {
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

func (cmd *CmdNet) links(term admin.Terminal, _ []string) error {
	var f = "%-8d %-24s %-8s %10s %10s %10s\n"

	links := cmd.mod.node.Network().Links().All()
	slices.SortFunc(links, func(a, b *network.ActiveLink) int {
		return cmp.Compare(a.ID(), b.ID())
	})

	term.Printf(f,
		admin.Header("ID"),
		admin.Header("Remote"),
		admin.Header("Net"),
		admin.Header("Idle"),
		admin.Header("Age"),
		admin.Header("Ping"),
	)
	for _, l := range links {
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
			l.RemoteIdentity(),
			admin.Keyword(net.Network(l)),
			idle,
			time.Since(l.AddedAt()).Round(time.Second),
			lat.Round(time.Millisecond),
		)
	}

	return nil
}

func (cmd *CmdNet) check(_ admin.Terminal, _ []string) error {
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

func (cmd *CmdNet) reroute(term admin.Terminal, args []string) error {
	if cmd.mod.relay == nil {
		return errModuleNotLoaded{"relay"}
	}

	corenode, ok := cmd.mod.node.(*node.CoreNode)
	if !ok {
		return errors.New("unsupported node type")
	}

	if len(args) < 2 {
		term.Printf("usage: net conn <ConnID> <LinkID>\n")
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

	lid, err := strconv.Atoi(args[1])
	if err != nil {
		return err
	}

	lnk, err := corenode.Network().Links().Find(lid)
	if err != nil {
		return err
	}

	return cmd.mod.relay.Reroute(conn.Query().Nonce(), lnk)
}

func (cmd *CmdNet) help(term admin.Terminal) error {
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

func selectEndpoints(list []net.Endpoint, selector func(net.Endpoint) bool) []net.Endpoint {
	var filtered = make([]net.Endpoint, 0)
	for _, e := range list {
		if selector(e) {
			filtered = append(filtered, e)
		}
	}
	return filtered
}
