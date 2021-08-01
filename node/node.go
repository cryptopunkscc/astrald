package node

import (
	"context"
	"github.com/cryptopunkscc/astrald/node/auth"
	"github.com/cryptopunkscc/astrald/node/auth/id"
	"github.com/cryptopunkscc/astrald/node/hub"
	"github.com/cryptopunkscc/astrald/node/link"
	"github.com/cryptopunkscc/astrald/node/net"
	"github.com/cryptopunkscc/astrald/node/net/etor"
	"github.com/cryptopunkscc/astrald/node/net/inet"
	"github.com/cryptopunkscc/astrald/node/net/lan"
	"github.com/cryptopunkscc/astrald/node/router"
	"io"
	"log"
)

type Node struct {
	*API
	Identity *id.ECIdentity
	TCPPort  int

	Hub    *hub.Hub
	Router *router.Router
}

func (node *Node) Connect(ctx context.Context, identity *id.ECIdentity, port string) (io.ReadWriteCloser, error) {
	// Establish a link with the identity
	l, err := node.Router.Connect(ctx, identity)
	if err != nil {
		return nil, net.ErrHostUnreachable
	}

	// Connect to identity's port
	return l.Open(port)
}

// New returns a new instance of a node
func New(identity *id.ECIdentity, port int) *Node {
	node := &Node{
		Identity: identity,
		TCPPort:  port,
		Hub:      new(hub.Hub),
		Router:   router.NewRouter(identity),
	}

	node.API = NewAPI(
		node.Identity,
		node.Router,
		node.Hub,
	)

	return node
}

// Run starts the node, waits for it to finish and returns an error if any
func (node *Node) Run(ctx context.Context) error {
	log.Printf("astral node '%s' starting...", node.Identity)

	newConnPipe := make(chan net.Conn)
	newLinkPipe := make(chan *link.Link)

	// Start listeners
	err := node.startListeners(ctx, newConnPipe)
	if err != nil {
		return err
	}

	// Start services
	for name, srv := range services {
		go func(name string, srv Service) {
			log.Printf("starting: %s...\n", name)
			err := srv.Run(ctx, node.API)
			if err != nil {
				log.Printf("error: %s: %v\n", name, err)
			} else {
				log.Printf("done: %s\n", name)
			}
		}(name, srv)
	}

	// Process incoming connections
	go node.processIncomingConns(newConnPipe, newLinkPipe)

	// Process incoming links
	go node.processNewLinks(newLinkPipe)

	// Wait for shutdown
	<-ctx.Done()
	close(newConnPipe)
	close(newLinkPipe)

	return nil
}

// startListeners starts listening to incoming connections
func (node *Node) startListeners(ctx context.Context, output chan<- net.Conn) error {
	net.Register(lan.NewDriver(node.Identity, uint16(node.TCPPort)))
	net.Register(inet.NewDriver())
	net.Register(etor.NewDriver())

	conns := net.Listen(ctx)

	// Start advertising
	net.Advertise(ctx)

	// Start scanning
	go func() {
		ads, err := net.Scan(ctx)
		if err != nil {
			log.Println("scan failed:", err)
			return
		}

		for ad := range ads {
			node.Router.Table.Add(ad.Identity.String(), ad.Addr)
		}
	}()

	// Start processing connections
	go func() {
		for conn := range conns {
			output <- conn
		}
	}()

	return nil
}

// processIncomingConns processes incoming connections from the input, upgrades them to a link and sends it to the output
func (node *Node) processIncomingConns(input <-chan net.Conn, output chan<- *link.Link) {
	for conn := range input {
		if conn.Outbound() {
			continue
		}

		log.Println(conn.RemoteAddr(), "connected.")
		authConn, err := auth.HandshakeInbound(context.Background(), conn, node.Identity)
		if err != nil {
			log.Println("handshake error:", err)
			_ = conn.Close()
			continue
		}

		output <- link.New(authConn)
	}
}

func (node *Node) processNewLinks(input <-chan *link.Link) {
	for link := range input {
		go node.handleLink(link)
	}
}

func (node *Node) handleLink(link *link.Link) {
	log.Println(link.RemoteIdentity(), "linked")
	node.Router.LinkCache.Add(link)
	for req := range link.Requests() {
		//Open a session with the service
		localStream, err := node.Hub.Connect(req.Port(), req.Caller())
		if err != nil {
			req.Reject()
			continue
		}

		// Accept remote party's request
		remoteStream, err := req.Accept()
		if err != nil {
			localStream.Close()
			continue
		}

		// Connect local and remote streams
		go func() {
			_, _ = io.Copy(localStream, remoteStream)
			_ = localStream.Close()
		}()
		go func() {
			_, _ = io.Copy(remoteStream, localStream)
			_ = remoteStream.Close()
		}()
	}
	node.Router.LinkCache.Remove(link.RemoteIdentity())
	log.Println(link.RemoteIdentity(), "unlinked")
}
