package node

import (
	"context"
	_id "github.com/cryptopunkscc/astrald/node/auth/id"
	_fs "github.com/cryptopunkscc/astrald/node/fs"
	"github.com/cryptopunkscc/astrald/node/hub"
	_ "github.com/cryptopunkscc/astrald/node/net/tcp"
	_ "github.com/cryptopunkscc/astrald/node/net/udp"
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
	PeerInfo *PeerInfo
	FS       *_fs.Filesystem
	Config   *Config
	Network  *Network
}

// New returns a new instance of a node
func New(astralDir string) *Node {
	fs := _fs.New(astralDir)
	identity := setupIdentity(fs)

	log.Printf("astral node %s statrting...", identity)

	peers := _peer.NewManager()
	peerInfo := NewPeerInfo(fs)

	node := &Node{
		FS:       fs,
		Identity: identity,
		Config:   loadConfig(fs),
		Hub:      new(hub.Hub),
		Peers:    peers,
		PeerInfo: peerInfo,
		Network:  NewNetwork(identity, peerInfo),
	}

	node.API = NewAPI(
		node.Identity,
		node.Peers,
		node.Hub,
		node.Network,
	)

	return node
}

// Run starts the node, waits for it to finish and returns an error if any
func (node *Node) Run(ctx context.Context) error {
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

	// Start network
	go func() {
		linkCh, _ := node.Network.Listen(ctx, node.Identity)

		for link := range linkCh {
			if err := node.Peers.AddLink(link); err != nil {
				log.Println("failed to add link:", err)
				link.Close()
			}
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

			link, err := node.Network.Link(nodeID)
			if err == nil {
				node.Peers.AddLink(link)
				break
			}
		}

		time.Sleep(5 * time.Second)
	}
}
