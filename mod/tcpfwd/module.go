package tcpfwd

import (
	"context"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/contacts"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/modules"
	"github.com/cryptopunkscc/astrald/streams"
	"net"
	"strings"
	"sync"
)

type Module struct {
	node   node.Node
	config Config
	log    *log.Logger
}

func (m *Module) Run(ctx context.Context) error {
	_, err := modules.WaitReady[*contacts.Module](ctx, m.node.Modules())
	if err != nil {
		return err
	}

	var wg sync.WaitGroup

	for astralPort, tcpPort := range m.config.Out {
		var astralPort = astralPort
		var tcpPort = tcpPort

		wg.Add(1)
		go func() {
			defer wg.Done()

			m.log.Logv(1, "forwarding %s to %s", astralPort, tcpPort)
			if err := m.ServeOut(ctx, astralPort, tcpPort); err != nil {
				m.log.Errorv(1, "error: %s", err)
			}
		}()
	}

	for tcpPort, astralPort := range m.config.In {
		var astralPort = astralPort
		var tcpPort = tcpPort

		wg.Add(1)
		go func() {
			defer wg.Done()

			err := m.ServeIn(ctx, tcpPort, astralPort)
			switch {
			case err == nil:
			case strings.Contains(err.Error(), "use of closed network connection"):
			default:
				m.log.Errorv(1, "error: %s", err)
			}
		}()
	}

	wg.Wait()
	return nil
}

func (m *Module) ServeOut(ctx context.Context, astral string, tcp string) error {
	port, err := m.node.Services().RegisterContext(ctx, astral, m.node.Identity())
	if err != nil {
		return err
	}

	for query := range port.Queries() {
		outConn, err := net.Dial("tcp", tcp)
		if err != nil {
			m.log.Errorv(1, "error forwarding %s to %s: %s", astral, tcp, err)
			query.Reject()
			continue
		}

		inConn, err := query.Accept()
		go streams.Join(inConn, outConn)
	}

	return nil
}

func (m *Module) ServeIn(ctx context.Context, tcp string, astral string) error {
	var nodeHex, port string
	var parts = strings.SplitN(astral, ":", 2)
	if len(parts) == 2 {
		nodeHex, port = parts[0], parts[1]
	} else {
		nodeHex, port = "localnode", parts[0]
	}

	nodeID, err := m.node.Resolver().Resolve(nodeHex)
	if err != nil {
		return err
	}

	listener, err := net.Listen("tcp", tcp)
	if err != nil {
		return err
	}

	go func() {
		<-ctx.Done()
		listener.Close()
	}()

	m.log.Logv(1, "forwarding %s to %s:%s", tcp, nodeID, port)

	for {
		inConn, err := listener.Accept()
		if err != nil {
			return err
		}

		outConn, err := m.node.Query(ctx, nodeID, port)
		if err != nil {
			inConn.Close()
			continue
		}

		go streams.Join(inConn, outConn)
	}
}
