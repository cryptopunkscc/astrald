package node

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/node/auth"
	"github.com/cryptopunkscc/astrald/node/auth/id"
	_link "github.com/cryptopunkscc/astrald/node/link"
	"github.com/cryptopunkscc/astrald/node/net"
)

type Network struct {
	identity *id.ECIdentity
	peerInfo *PeerInfo
}

func NewNetwork(identity *id.ECIdentity, peerInfo *PeerInfo) *Network {
	return &Network{identity: identity, peerInfo: peerInfo}
}

func (n *Network) Link(remoteID *id.ECIdentity) (*_link.Link, error) {
	// Get node endpoints
	endpoints := n.peerInfo.Find(remoteID.String())
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

func (n *Network) Listen(ctx context.Context, identity *id.ECIdentity) (<-chan *_link.Link, <-chan error) {
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

func (n *Network) Advertise(ctx context.Context, identity *id.ECIdentity) error {
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
