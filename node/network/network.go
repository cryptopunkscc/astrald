package network

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node/event"
	"github.com/cryptopunkscc/astrald/node/link"
)

type Network interface {
	Link(context.Context, id.Identity) (*link.Link, error)
	Events() *event.Queue
	Peers() *PeerSet
	Server() *Server
	AddLink(*link.Link) error
	AddSecureConn(conn net.SecureConn) error
}
