package node

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/astral/link"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/hub"
	"github.com/cryptopunkscc/astrald/logfmt"
	"github.com/cryptopunkscc/astrald/node/network"
	"github.com/cryptopunkscc/astrald/node/storage"
	"io"
	"log"
	"time"
)

type Node struct {
	Config   *Config
	Identity id.Identity
	Ports    *hub.Hub
	Network  *network.Network
	Store    storage.Store

	futureEvent *FutureEvent
}

// New returns a new instance of a node
func New(astralDir string) *Node {
	fs := storage.NewFilesystemStorage(astralDir)

	identity := setupIdentity(fs)

	log.Printf("astral node %s statrting...", identity)

	config := loadConfig(fs)

	node := &Node{
		Store:       fs,
		Identity:    identity,
		Config:      config,
		Ports:       hub.New(),
		Network:     network.NewNetwork(config.Network, identity, fs),
		futureEvent: NewFutureEvent(),
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

	// start the network
	reqCh, eventCh, reqErrCh := node.Network.Run(ctx, node.Identity)

	go func() {
		time.Sleep(2 * time.Second)
		log.Println("public info", node.Network.Info(true))
	}()

	for {
		select {
		case request := <-reqCh:
			node.handleRequest(request)

		case event := <-eventCh:
			node.handleEvent(event)

		case err := <-reqErrCh:
			log.Println("fatal error:", err)
			return err

		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (node *Node) Query(ctx context.Context, remoteID id.Identity, query string) (io.ReadWriteCloser, error) {
	if remoteID.IsEmpty() || remoteID.IsEqual(node.Identity) {
		return node.Ports.Query(query, node.Identity)
	}

	return node.Network.Query(ctx, remoteID, query)
}

func (node *Node) ResolveIdentity(str string) (id.Identity, error) {
	if id, err := id.ParsePublicKeyHex(str); err == nil {
		return id, nil
	}

	target, found := node.Config.Alias[str]

	if !found {
		return id.Identity{}, errors.New("unknown identity")
	}
	if str == target {
		return id.Identity{}, errors.New("circular alias")
	}
	return node.ResolveIdentity(target)
}

func (node *Node) FutureEvent() *FutureEvent {
	return node.futureEvent
}

func (node *Node) handleRequest(request link.Request) {
	if request.Query() == ".ping" {
		log.Println("ping from", logfmt.ID(request.Caller()))
		request.Reject()
		return
	}

	// Query a session with the service
	localStream, err := node.Ports.Query(request.Query(), request.Caller())
	if err != nil {
		request.Reject()
		log.Printf("%s rejected %s\n", logfmt.ID(request.Caller()), request.Query())
		return
	}

	// Accept remote party's request
	remoteStream, err := request.Accept()
	if err != nil {
		localStream.Close()
		return
	}

	log.Printf("%s accepted %s\n", logfmt.ID(request.Caller()), request.Query())

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

func (node *Node) handleEvent(event Eventer) {
	switch event := event.(type) {
	case network.Event:
		node.handleNetworkEvent(event)
	}

	node.futureEvent = node.futureEvent.done(event)
}

func (node *Node) handleNetworkEvent(event network.Event) {
	switch event.Event() {
	case network.EventPeerLinked:
		log.Println(logfmt.ID(event.Peer.Identity()), "linked")
	case network.EventPeerUnlinked:
		log.Println(logfmt.ID(event.Peer.Identity()), "unlinked")
	}
}
