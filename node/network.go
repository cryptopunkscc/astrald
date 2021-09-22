package node

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/auth"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/net"
	_link "github.com/cryptopunkscc/astrald/node/link"
)

type Network struct {
	identity *id.Identity
	peerInfo *PeerInfo
}

func (n *Network) Identity() *id.Identity {
	return n.identity
}

func NewNetwork(identity *id.Identity, peerInfo *PeerInfo) *Network {
	return &Network{
		identity: identity,
		peerInfo: peerInfo,
	}
}

func (n *Network) LinkAt(remoteID *id.Identity, addr net.Addr) (*_link.Link, error) {
	netConn, err := net.Dial(context.Background(), addr)
	if err != nil {
		return nil, err
	}

	authConn, err := auth.HandshakeOutbound(context.Background(), netConn, remoteID, n.identity)
	if err != nil {
		return nil, err
	}

	return _link.New(authConn), nil
}

func (n *Network) Link(remoteID *id.Identity) (*_link.Link, error) {
	// Get node endpoints
	endpoints := n.peerInfo.NodeAddr(remoteID.String())
	if len(endpoints) == 0 {
		return nil, errors.New("no endpoints found for the host")
	}

	// Establish a connection
	netConn, err := n.tryEndpoints(endpoints)
	if err != nil {
		return nil, err
	}

	// Authenticate via handshake
	authConn, err := auth.HandshakeOutbound(context.Background(), netConn, remoteID, n.identity)
	if err != nil {
		return nil, err
	}

	link := _link.New(authConn)

	return link, nil
}

func (n *Network) Listen(ctx context.Context, identity *id.Identity) (<-chan *_link.Link, <-chan error) {
	linkCh := make(chan *_link.Link)
	errCh := make(chan error)

	go func() {
		defer close(linkCh)

		connCh := net.Listen(ctx)

		for conn := range connCh {
			authConn, err := auth.HandshakeInbound(ctx, conn, identity)
			if err != nil {
				conn.Close()
				continue
			}

			linkCh <- _link.New(authConn)
		}
	}()

	return linkCh, errCh
}

func (n *Network) Advertise(ctx context.Context, identity *id.Identity) error {
	return net.Advertise(ctx, identity.String())
}

func (n *Network) Scan(ctx context.Context) (<-chan *net.Ad, error) {
	return net.Scan(ctx)
}

func (n *Network) tryEndpoints(endpoints []net.Addr) (net.Conn, error) {
	for _, endpoint := range endpoints {
		// Establish a connection
		netConn, err := net.Dial(context.Background(), endpoint)
		if err == nil {
			return netConn, nil
		}
	}
	return nil, errors.New("no endpoint could be reached")
}
