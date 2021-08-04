package router

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/node/auth"
	"github.com/cryptopunkscc/astrald/node/auth/id"
	_fs "github.com/cryptopunkscc/astrald/node/fs"
	"github.com/cryptopunkscc/astrald/node/link"
	"github.com/cryptopunkscc/astrald/node/net"
)

// Router keeps track of visibility and links to other nodes
type Router struct {
	Table         *Table
	LinkCache     *LinkCache
	localIdentity *id.ECIdentity
}

func NewRouter(fs *_fs.Filesystem, localID *id.ECIdentity) *Router {
	return &Router{
		Table:         NewTable(fs),
		LinkCache:     NewLinkCache(),
		localIdentity: localID,
	}
}

func (router *Router) Connect(ctx context.Context, remoteID *id.ECIdentity) (*link.Link, error) {
	// Check if we already have a link first
	if l := router.LinkCache.Fetch(remoteID); l != nil {
		return l, nil
	}

	// Get node endpoints
	endpoints := router.Table.Find(remoteID.String())
	if len(endpoints) == 0 {
		return nil, errors.New("no endpoints found for the host")
	}

	// Establish a connection
	netConn, err := router.tryEndpoints(ctx, endpoints)
	if err != nil {
		return nil, err
	}

	// Authenticate via handshake
	authConn, err := auth.HandshakeOutbound(ctx, netConn, remoteID, router.localIdentity)
	if err != nil {
		return nil, err
	}

	l := link.New(authConn)

	router.LinkCache.Add(l)

	return l, nil
}

func (router *Router) tryEndpoints(ctx context.Context, endpoints []net.Addr) (net.Conn, error) {
	for _, endpoint := range endpoints {
		// Establish a connection
		netConn, err := net.Dial(ctx, endpoint)
		if err == nil {
			return netConn, nil
		}
	}
	return nil, errors.New("no endpoint could be reached")
}
