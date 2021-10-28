package node

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/link"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/hub"
	"github.com/cryptopunkscc/astrald/infra"
	"github.com/cryptopunkscc/astrald/logfmt"
	"github.com/cryptopunkscc/astrald/node/fs"
	"github.com/cryptopunkscc/astrald/node/linker"
	"github.com/cryptopunkscc/astrald/node/network"
	"github.com/cryptopunkscc/astrald/node/route"
	"io"
	"log"
	"time"
)

const defaultLinkTimeout = 30 * time.Second

type Node struct {
	Identity id.Identity
	Ports    *hub.Hub
	Network  *network.Network
	Linker   *linker.Linker
	Router   *BasicRouter
	Config   *Config
	FS       *fs.Filesystem
}

// New returns a new instance of a node
func New(astralDir string) *Node {
	fs := fs.New(astralDir)

	identity := setupIdentity(fs)

	log.Printf("astral node %s statrting...", identity)

	hub := hub.New()
	config := loadConfig(fs)

	router := NewBasicRouter()
	node := &Node{
		FS:       fs,
		Identity: identity,
		Config:   config,
		Ports:    hub,
		Router:   router,
		Network:  network.NewNetwork(config.Network),
	}

	node.loadRoutes()

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
	requests, requestsErr := node.Network.Run(ctx, node.Identity)

	node.Linker = linker.NewLinker(ctx, node.Identity, node.Router)

	go func() {
		time.Sleep(2 * time.Second)
		log.Println("public route", node.Route(true))
	}()

	defer node.saveRoutes()

	for {
		select {
		case request := <-requests:
			if err := node.handleRequest(request); err != nil {
				log.Println("error handling request:", err)
			}

		case link := <-node.Linker.Links():
			node.Network.AddLink(link)

		case err := <-requestsErr:
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

	link, err := node.Link(ctx, remoteID)
	if err != nil {
		return nil, err
	}

	return link.Query(query)
}

func (node *Node) Link(ctx context.Context, remoteID id.Identity) (*link.Link, error) {
	peer := node.Network.View.Peer(remoteID)

	node.Linker.Wake(remoteID)

	select {
	case <-peer.WaitLinked():
		return peer.PreferredLink(), nil
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(defaultLinkTimeout):
		return nil, errors.New("timeout")
	}
}

func (node *Node) Route(onlyPublic bool) *route.Route {
	addrs := make([]infra.Addr, 0)

	for _, a := range astral.Addresses() {
		if onlyPublic && !a.Public {
			continue
		}
		addrs = append(addrs, a.Addr)
	}

	return &route.Route{
		Identity:  node.Identity,
		Addresses: addrs,
	}
}

func (node *Node) handleRequest(request link.Request) error {
	if request.Query() == ".ping" {
		log.Println("ping from", logfmt.ID(request.Caller()))
		return request.Reject()
	}

	//Query a session with the service
	localStream, err := node.Ports.Query(request.Query(), request.Caller())
	if err != nil {
		request.Reject()
		return err
	}

	// Accept remote party's request
	remoteStream, err := request.Accept()
	if err != nil {
		localStream.Close()
		return err
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

	return nil
}

func (node *Node) loadRoutes() error {
	data, err := node.FS.Read("routes")
	if err != nil {
		return err
	}

	return node.Router.AddPacked(data)
}

func (node *Node) saveRoutes() {
	err := node.FS.Write("routes", node.Router.Pack())
	if err != nil {
		log.Println("save routes error:", err)
	}
}
