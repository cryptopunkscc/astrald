package node

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/node/auth"
	"github.com/cryptopunkscc/astrald/node/auth/id"
	_link "github.com/cryptopunkscc/astrald/node/link"
	"github.com/cryptopunkscc/astrald/node/net"
)

type Linker struct {
	peerInfo      *PeerInfo
	localIdentity *id.ECIdentity
}

func NewLinker(peerInfo *PeerInfo, localIdentity *id.ECIdentity) *Linker {
	return &Linker{peerInfo: peerInfo, localIdentity: localIdentity}
}

func (l *Linker) Link(remoteID *id.ECIdentity) (*_link.Link, error) {
	// Get node endpoints
	endpoints := l.peerInfo.Find(remoteID.String())
	if len(endpoints) == 0 {
		return nil, errors.New("no endpoints found for the host")
	}

	// Establish a connection
	netConn, err := l.tryEndpoints(endpoints)
	if err != nil {
		return nil, err
	}

	// Authenticate via handshake
	authConn, err := auth.HandshakeOutbound(context.Background(), netConn, remoteID, l.localIdentity)
	if err != nil {
		return nil, err
	}

	link := _link.New(authConn)

	return link, nil
}

func (l *Linker) tryEndpoints(endpoints []net.Addr) (net.Conn, error) {
	for _, endpoint := range endpoints {
		// Establish a connection
		netConn, err := net.Dial(context.Background(), endpoint)
		if err == nil {
			return netConn, nil
		}
	}
	return nil, errors.New("no endpoint could be reached")
}
