package node

import (
	"bytes"
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/link"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/hub"
	"github.com/cryptopunkscc/astrald/infra"
	"github.com/cryptopunkscc/astrald/infra/inet"
	"github.com/cryptopunkscc/astrald/infra/tor"
	"github.com/cryptopunkscc/astrald/logfmt"
	"github.com/cryptopunkscc/astrald/node/fs"
	"github.com/cryptopunkscc/astrald/node/route"
	"io"
	"log"
	"time"
)

const defaultLinkTimeout = 30 * time.Second

type Node struct {
	Identity id.Identity
	Ports    *hub.Hub
	Links    *link.Set
	Routes   map[string]*route.Route

	Config       *Config
	FS           *fs.Filesystem
	LinkerStatus map[string]struct{}
	requests     chan link.Request
}

// New returns a new instance of a node
func New(astralDir string) *Node {
	_inet := inet.New()
	err := astral.AddNetwork(_inet)
	if err != nil {
		log.Println("error adding inet:", err)
	}

	err = astral.AddNetwork(tor.New())
	if err != nil {
		log.Println("error adding tor:", err)
	}

	fs := fs.New(astralDir)
	identity := setupIdentity(fs)

	log.Printf("astral node %s statrting...", identity)

	hub := hub.New()

	node := &Node{
		FS:           fs,
		Identity:     identity,
		Config:       loadConfig(fs),
		Ports:        hub,
		Links:        link.NewSet(),
		LinkerStatus: make(map[string]struct{}),
		requests:     make(chan link.Request),
		Routes:       make(map[string]*route.Route),
	}

	if node.Config.ExternalAddr != "" {
		err := _inet.AddExternalAddr(node.Config.ExternalAddr)
		if err != nil {
			log.Println("external ip error:", err)
		}
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

	links, _ := astral.Listen(ctx, node.Identity)

	defer node.saveRoutes()

	go func() {
		time.Sleep(2 * time.Second)
		log.Println("public route", node.Route(true))
	}()

	for {
		select {
		case link := <-links:
			if err := node.addLink(link); err != nil {
				log.Println("error adding link:", err)
				link.Close()
			}
		case request := <-node.requests:
			if err := node.handleRequest(request); err != nil {
				log.Println("error handling request:", err)
			}
		case <-ctx.Done():
			return nil
		}
	}
}

func (node *Node) saveRoutes() {
	buf := &bytes.Buffer{}
	for _, r := range node.Routes {
		_ = route.Write(buf, r)
	}
	err := node.FS.Write("routes", buf.Bytes())
	if err != nil {
		log.Println("save routes error:", err)
	}
}

func (node *Node) loadRoutes() {
	data, err := node.FS.Read("routes")
	if err != nil {
		log.Println("load routes error:", err)
		return
	}

	buf := bytes.NewReader(data)
	for {
		r, err := route.Read(buf)
		if err != nil {
			if !errors.Is(err, io.EOF) {
				log.Println("load routes error:", err)
			} else {
				log.Println("loaded", len(node.Routes), "routes")
			}
			return
		}
		node.AddRoute(r)
	}
}

func (node *Node) Query(remoteID id.Identity, query string) (io.ReadWriteCloser, error) {
	if remoteID.IsEmpty() || remoteID.IsEqual(node.Identity) {
		return node.Ports.Query(query, node.Identity)
	}

	link, err := node.Link(remoteID)
	if err != nil {
		return nil, err
	}

	return link.Query(query)
}

func (node *Node) RequestLink(remoteID id.Identity) {
	pkh := remoteID.PublicKeyHex()

	if _, found := node.LinkerStatus[pkh]; found {
		return
	}

	node.LinkerStatus[pkh] = struct{}{}

	go func() {
		linker := NewLinker(node.Identity, remoteID, node.Routes)
		for lnk := range linker.Run(context.Background()) {
			err := node.addLink(lnk)
			if err != nil {
				log.Println("add link error:", err)
			}
		}
		delete(node.LinkerStatus, pkh)
	}()
}

func (node *Node) Link(remoteID id.Identity) (*link.Link, error) {
	return node.LinkWithTimeout(remoteID, defaultLinkTimeout)
}

func (node *Node) LinkWithTimeout(remoteID id.Identity, timeout time.Duration) (*link.Link, error) {
	links, cancel := node.Links.Watch(true)
	peer := links.Peer(remoteID)
	defer cancel()

	select {
	case l := <-peer:
		return l, nil
	case <-time.After(1 * time.Millisecond):
		node.RequestLink(remoteID)
	}

	select {
	case l := <-peer:
		return l, nil
	case <-time.After(timeout):
		return nil, errors.New("link timeout")
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

func (node *Node) addLink(lnk *link.Link) error {
	err := node.Links.Add(lnk)
	if err != nil {
		return err
	}

	addr, nodeID := lnk.RemoteAddr(), lnk.RemoteIdentity()

	if node.Links.All().Peer(nodeID).Count() == 1 {
		log.Println("linked", logfmt.ID(lnk.RemoteIdentity()))
	}

	log.Println("link up", logfmt.ID(nodeID), "via", addr.Network(), addr.String())

	go func() {
		for req := range lnk.Requests() {
			node.requests <- req
		}
		log.Println("link down", logfmt.ID(nodeID), "via", addr.Network(), addr.String())

		if node.Links.All().Peer(nodeID).Count() == 0 {
			log.Println("unlinked", logfmt.ID(lnk.RemoteIdentity()))
		}
	}()

	return nil
}

func (node *Node) AddRoute(r *route.Route) {
	hex := r.Identity.PublicKeyHex()

	if _, found := node.Routes[hex]; !found {
		node.Routes[hex] = route.New(r.Identity)
	}

	_route := node.Routes[hex]
	for _, a := range r.Addresses {
		_route.Add(a)
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
