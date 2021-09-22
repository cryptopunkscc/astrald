package node

import (
	"context"
	_id "github.com/cryptopunkscc/astrald/auth/id"
	_hub "github.com/cryptopunkscc/astrald/hub"
	_ "github.com/cryptopunkscc/astrald/net/tcp"
	_ "github.com/cryptopunkscc/astrald/net/tor"
	_ "github.com/cryptopunkscc/astrald/net/udp"
	_fs "github.com/cryptopunkscc/astrald/node/fs"
	_peer "github.com/cryptopunkscc/astrald/node/peer"
	"io"
	"log"
)

type Node struct {
	Identity *_id.Identity

	Hub       *_hub.Hub
	Peers     *_peer.Peers
	PeerInfo  *PeerInfo
	FS        *_fs.Filesystem
	Config    *Config
	Network   *Network
	LinkCache *LinkCache
}

// New returns a new instance of a node
func New(astralDir string) *Node {
	fs := _fs.New(astralDir)
	identity := setupIdentity(fs)

	log.Printf("astral node %s statrting...", identity)

	peers := _peer.NewManager()
	peerInfo := NewPeerInfo(fs)
	hub := _hub.NewHub()

	node := &Node{
		FS:        fs,
		Identity:  identity,
		Config:    loadConfig(fs),
		Hub:       hub,
		Peers:     peers,
		PeerInfo:  peerInfo,
		Network:   NewNetwork(identity, peerInfo),
		LinkCache: NewLinkCache(),
	}

	return node
}

// Run starts the node, waits for it to finish and returns an error if any
func (node *Node) Run(ctx context.Context) error {
	// Start services
	for name, srv := range services {
		go func(name string, srv ServiceRunner) {
			log.Printf("starting %s...\n", name)
			err := srv(ctx, node)
			if err != nil {
				log.Printf("%s failed: %v\n", name, err)
			} else {
				log.Printf("%s done.\n", name)
			}
		}(name, srv)
	}

	// Start network
	go func() {
		linkCh, _ := node.Network.Listen(ctx, node.Identity)

		for link := range linkCh {
			if err := node.Peers.AddLink(link); err != nil {
				log.Println("failed to add link:", err)
				link.Close()
			}
			node.LinkCache.AddLink(link)
		}
	}()

	go node.Network.Advertise(ctx, node.Identity)

	go func() {
		adCh, err := node.Network.Scan(ctx)
		if err != nil {
			log.Println("scan error:", err)
			return
		}

		for ad := range adCh {
			if ad.Identity.String() == node.Identity.String() {
				continue
			}
			node.PeerInfo.Add(ad.Identity.String(), ad.Addr)
		}
	}()

	// Handle incoming network requests
	go node.handleRequests()

	// Keep alive LAN peers
	go node.linkKeeper()

	// Wait for shutdown
	<-ctx.Done()

	return nil
}

func (node *Node) Connect(nodeID *_id.Identity, query string) (io.ReadWriteCloser, error) {
	if (nodeID == nil) || (nodeID.String() == node.Identity.String()) {
		return node.Hub.Connect(query, node.Identity)
	}

	peer, _ := node.Peers.Peer(nodeID)

	if !peer.Connected() {
		link, err := node.Network.Link(nodeID)
		if err != nil {
			return nil, err
		}

		node.Peers.AddLink(link)
		node.LinkCache.AddLink(link)
	}

	return peer.DefaultLink().Query(query)
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
	for keyHex := range node.PeerInfo.entries {
		nodeID, _ := _id.ParsePublicKeyHex(keyHex)

		go func() {
			for lnk := range PeerLink(nodeID, node.PeerInfo, node.Network) {
				err := node.Peers.AddLink(lnk)
				if err != nil {
					log.Println("add link error:", err)
				}
				err = node.LinkCache.AddLink(lnk)
				if err != nil {
					log.Println("add link error:", err)
				}
			}
		}()
	}
}
