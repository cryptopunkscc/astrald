package admin

import (
	"errors"
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/nodes/src/muxlink"
	"github.com/cryptopunkscc/astrald/net"
	"reflect"
	"strconv"
	"strings"
	"time"
)

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
	case "conns":
		return cmd.conns(term, args[2:])

	case "conn":
		return cmd.conn(term, args[2:])

	case "help":
		return cmd.help(term)

	default:
		return errors.New("invalid command")
	}
}

func (cmd *CmdNet) unlinkHelp(term admin.Terminal) error {
	term.Printf("help: net unlink <node>\n\n")
	return nil
}

func (cmd *CmdNet) conns(term admin.Terminal, _ []string) error {
	corenode, ok := cmd.mod.node.(*core.CoreNode)
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
	corenode, ok := cmd.mod.node.(*core.CoreNode)
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
		case *muxlink.PortBinding:
			term.Printf("  Identity: %d\n", w.Output().Identity())
			term.Printf("  Port: %d\n", w.Port())
			if t := w.Transport().(exonet.Conn); t != nil {
				var (
					network                       = "unknown"
					localEndpoint, remoteEndpoint string
				)
				if e := t.LocalEndpoint(); e != nil {
					network = e.Network()
					localEndpoint = e.Address()
				}

				if e := t.RemoteEndpoint(); e != nil {
					remoteEndpoint = e.Address()
				}

				term.Printf("  Transport: %s %s~%s\n",
					network,
					localEndpoint,
					remoteEndpoint,
				)
			}
			term.Printf("  Buffer: %d/%d\n", w.Used(), w.BufferSize())

		case *muxlink.PortWriter:
			term.Printf("  Identity: %d\n", w.Identity())
			term.Printf("  Port: %d\n", w.Port())
			term.Printf("  Source: %s\n", reflect.TypeOf(w.Source()))
			if t := w.Transport().(exonet.Conn); t != nil {
				var (
					network                       = "unknown"
					localEndpoint, remoteEndpoint string
				)
				if e := t.LocalEndpoint(); e != nil {
					network = e.Network()
					localEndpoint = e.Address()
				}

				if e := t.RemoteEndpoint(); e != nil {
					remoteEndpoint = e.Address()
				}

				term.Printf("  Transport: %s %s~%s\n",
					network,
					localEndpoint,
					remoteEndpoint,
				)
			}
			term.Printf("  Buffer: %d\n", w.BufferSize())

		case *core.MonitoredWriter:
			term.Printf("  Identity: %d\n", w.Identity())
			term.Printf("  Bytes: %d\n", w.Bytes())

		case *net.SecurePipeWriter:
			term.Printf("  Identity: %d\n", w.Identity())
			term.Printf("  Transport: %s\n", reflect.TypeOf(w.Transport()))

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

func selectEndpoints(list []exonet.Endpoint, selector func(exonet.Endpoint) bool) []exonet.Endpoint {
	var filtered = make([]exonet.Endpoint, 0)
	for _, e := range list {
		if selector(e) {
			filtered = append(filtered, e)
		}
	}
	return filtered
}
