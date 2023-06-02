package network

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node/events"
	"github.com/cryptopunkscc/astrald/node/link"
)

type Network interface {
	Link(context.Context, id.Identity) (*link.Link, error)
	Query(context.Context, id.Identity, string) (link.BasicConn, error)
	Events() *events.Queue
	Peers() *PeerSet
	Server() *Server
	AddLink(*link.Link) error
	AddSecureConn(conn net.SecureConn) error
}
