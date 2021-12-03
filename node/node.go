package node

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/hub"
	"github.com/cryptopunkscc/astrald/logfmt"
	"github.com/cryptopunkscc/astrald/node/link"
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
	queryCh, eventCh, queryErrCh := node.Network.Run(ctx, node.Identity)

	go func() {
		time.Sleep(2 * time.Second)
		log.Println("public info", node.Network.Info(true))
	}()

	for {
		select {
		case query := <-queryCh:
			go node.serveQuery(query)

		case event := <-eventCh:
			node.handleEvent(event)

		case err := <-queryErrCh:
			log.Println("fatal error:", err)
			return err

		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (node *Node) Query(ctx context.Context, remoteID id.Identity, query string) (io.ReadWriteCloser, error) {
	if remoteID.IsZero() || remoteID.IsEqual(node.Identity) {
		return node.Ports.Query(query, node.Identity)
	}

	return node.Network.Query(ctx, remoteID, query)
}

func (node *Node) ResolveIdentity(str string) (id.Identity, error) {
	if id, err := id.ParsePublicKeyHex(str); err == nil {
		return id, nil
	}

	target, found := node.Network.Contacts.ResolveAlias(str)

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

func (node *Node) serveQuery(query *link.Query) {
	if query.String() == ".ping" {
		log.Println("ping from", logfmt.ID(query.Caller()))
		query.Reject()
		return
	}

	log.Printf("<- [%s]:%s (%s)\n", logfmt.ID(query.Caller()), query, query.Link().Network())

	// Query a session with the service
	localStream, err := node.Ports.Query(query.String(), query.Caller())
	if err != nil {
		query.Reject()
		log.Printf("%s rejected %s\n", logfmt.ID(query.Caller()), query.String())
		return
	}

	// Accept remote party's query
	remoteStream, err := query.Accept()
	if err != nil {
		localStream.Close()
		return
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
