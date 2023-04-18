package tcpfwd

import (
	"context"
	_log "github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/streams"
	"net"
	"strings"
	"sync"
)

var log = _log.Tag(ModuleName)

type Module struct {
	node   node.Node
	config Config
}

func (m *Module) Run(ctx context.Context) error {
	var wg sync.WaitGroup

	for astralPort, tcpPort := range m.config.Out {
		var astralPort = astralPort
		var tcpPort = tcpPort

		wg.Add(1)
		go func() {
			defer wg.Done()

			log.Logv(1, "forwarding %s to %s", astralPort, tcpPort)
			if err := m.ServeOut(ctx, astralPort, tcpPort); err != nil {
				log.Errorv(1, "error: %s", err)
			}
		}()
	}

	for tcpPort, astralPort := range m.config.In {
		var astralPort = astralPort
		var tcpPort = tcpPort

		wg.Add(1)
		go func() {
			defer wg.Done()

			log.Logv(1, "forwarding %s to %s", tcpPort, astralPort)
			if err := m.ServeIn(ctx, tcpPort, astralPort); err != nil {
				log.Errorv(1, "error: %s", err)
			}
		}()
	}

	wg.Wait()
	return nil
}

func (m *Module) ServeOut(ctx context.Context, astral string, tcp string) error {
	port, err := m.node.Services().RegisterContext(ctx, astral)
	if err != nil {
		return err
	}

	for query := range port.Queries() {
		outConn, err := net.Dial("tcp", tcp)
		if err != nil {
			log.Errorv(1, "error forwarding %s to %s: %s", astral, tcp, err)
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

	nodeID, err := m.node.Contacts().ResolveIdentity(nodeHex)
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