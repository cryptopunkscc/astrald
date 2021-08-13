package node

import (
	"context"
	"github.com/cryptopunkscc/astrald/node/auth"
	_id "github.com/cryptopunkscc/astrald/node/auth/id"
	_fs "github.com/cryptopunkscc/astrald/node/fs"
	"github.com/cryptopunkscc/astrald/node/hub"
	_link "github.com/cryptopunkscc/astrald/node/link"
	"github.com/cryptopunkscc/astrald/node/net"
	"github.com/cryptopunkscc/astrald/node/net/inet"
	"github.com/cryptopunkscc/astrald/node/net/lan"
	_peer "github.com/cryptopunkscc/astrald/node/peer"
	"io"
	"log"
	"time"
)

type Node struct {
	*API
	Identity *_id.ECIdentity

	Hub      *hub.Hub
	Peers    *_peer.Peers
	Linker   *Linker
	PeerInfo *PeerInfo
	FS       *_fs.Filesystem
	Config   *Config
}

// New returns a new instance of a node
func New(astralDir string) *Node {
	fs := _fs.New(astralDir)
	identity := setupIdentity(fs)
	peers := _peer.NewManager()
	peerInfo := NewPeerInfo(fs)

	node := &Node{
		FS:       fs,
		Identity: identity,
		Config:   loadConfig(fs),
		Hub:      new(hub.Hub),
		Peers:    peers,
		PeerInfo: peerInfo,
		Linker:   NewLinker(peerInfo, identity),
	}

	node.API = NewAPI(
		node.Identity,
		node.Peers,
		node.Hub,
		node.Linker,
	)

	return node
}

// Run starts the node, waits for it to finish and returns an error if any
func (node *Node) Run(ctx context.Context) error {
	log.Printf("astral node '%s' starting...", node.Identity)

	newConnPipe := make(chan net.Conn)
	newLinkPipe := make(chan *_link.Link)

	// Start listeners
	err := node.startListeners(ctx, newConnPipe)
	if err != nil {
		return err
	}

	// Start services
	for name, srv := range services {
		go func(name string, srv ServiceRunner) {
			log.Printf("starting %s...\n", name)
			err := srv(ctx, node.API)
			if err != nil {
				log.Printf("%s failed: %v\n", name, err)
			} else {
				log.Printf("%s done.\n", name)
			}
		}(name, srv)
	}

	// Process incoming connections
	go node.processIncomingConns(newConnPipe, newLinkPipe)

	// Process incoming links
	go node.processNewLinks(newLinkPipe)

	// Handle incoming network requests
	go node.handleRequests()

	// Keep alive LAN peers
	go node.linkKeeper()

	// Wait for shutdown
	<-ctx.Done()
	close(newConnPipe)
	close(newLinkPipe)

	return nil
}

// startListeners starts listening to incoming connections
func (node *Node) startListeners(ctx context.Context, output chan<- net.Conn) error {
	net.Register(lan.NewDriver(node.Identity, uint16(node.Config.Port)))
	net.Register(inet.NewDriver())
	//net.Register(etor.NewDriver())

	conns := net.Listen(ctx)

	// Start advertising
	// net.Advertise(ctx)

	// Start scanning
	go func() {
		ads, err := net.Scan(ctx)
		if err != nil {
			log.Println("scan failed:", err)
			return
		}

		for ad := range ads {
			node.PeerInfo.Add(ad.Identity.String(), ad.Addr)
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
func (node *Node) processIncomingConns(input <-chan net.Conn, output chan<- *_link.Link) {
	for conn := range input {
		if conn.Outbound() {
			continue
		}

		log.Println("new connection from", conn.RemoteAddr())
		authConn, err := auth.HandshakeInbound(context.Background(), conn, node.Identity)
		if err != nil {
			log.Println("handshake error:", err)
			_ = conn.Close()
			continue
		}

		output <- _link.New(authConn)
	}
}

func (node *Node) processNewLinks(input <-chan *_link.Link) {
	for link := range input {
		node.Peers.AddLink(link)
	}
}

func (node *Node) handleRequests() {
	for req := range node.Peers.Requests() {
		//Query a session with the service
		localStream, err := node.Hub.Connect(req.Query(), req.Caller())
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
}

// linkKeeper tries to keep us connected to all peers in our local database
func (node *Node) linkKeeper() {
	for {
		for key := range node.PeerInfo.entries {
			nodeID, _ := _id.ParsePublicKeyHex(key)
			peer, _ := node.Peers.Peer(nodeID)

			if peer.Connected() {
				continue
			}

			link, err := node.Linker.Link(nodeID)
			if err == nil {
				node.Peers.AddLink(link)
				break
			}
		}

		time.Sleep(5 * time.Second)
	}
}
