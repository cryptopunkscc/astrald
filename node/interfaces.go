package node

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/hub"
	"github.com/cryptopunkscc/astrald/infra"
	"github.com/cryptopunkscc/astrald/infra/inet"
	"github.com/cryptopunkscc/astrald/node/config"
	"github.com/cryptopunkscc/astrald/node/contacts"
	"github.com/cryptopunkscc/astrald/node/event"
	"github.com/cryptopunkscc/astrald/node/link"
	"github.com/cryptopunkscc/astrald/node/network"
	"github.com/cryptopunkscc/astrald/node/tracker"
	"time"
)

type Node interface {
	Identity() id.Identity
	Alias() string
	SetAlias(alias string) error
	Query(ctx context.Context, remoteID id.Identity, query string) (link.BasicConn, error)
	RootDir() string
	ConfigStore() config.Store
	Events() *event.Queue
	Infra() Infra
	Network() Network
	Tracker() Tracker
	Contacts() Contacts
	Services() Services
}

// Infra is an interface for infrastructural networks
type Infra interface {
	Dial(ctx context.Context, addr infra.Addr) (conn infra.Conn, err error)
	LocalAddrs() []infra.AddrSpec
	Unpack(network string, data []byte) (infra.Addr, error)
	Inet() *inet.Inet
}

type Network interface {
	Link(ctx context.Context, nodeID id.Identity) (*link.Link, error)
	Events() *event.Queue
	AddLink(l *link.Link) error
	Peers() *network.PeerSet
	Server() *network.Server
	AddAuthConn(conn auth.Conn) error
}

type Tracker interface {
	Add(identity id.Identity, addr infra.Addr, expiresAt time.Time) error
	Identities() ([]id.Identity, error)
	ForgetIdentity(identity id.Identity) error
	Watch(ctx context.Context, nodeID id.Identity) <-chan *tracker.Addr
	AddrByIdentity(identity id.Identity) ([]tracker.Addr, error)
}

type Contacts interface {
	DisplayName(nodeID id.Identity) string
	Find(nodeID id.Identity) (c *contacts.Contact, err error)
	FindOrCreate(nodeID id.Identity) (c *contacts.Contact, err error)
	FindByAlias(alias string) (c *contacts.Contact, found bool)
	Delete(identity id.Identity) error
	ResolveIdentity(str string) (id.Identity, error)
	All() <-chan *contacts.Contact
}

type Services interface {
	Register(name string) (*hub.Port, error)
	RegisterContext(ctx context.Context, name string) (*hub.Port, error)
	Query(ctx context.Context, query string, link *link.Link) (*hub.Conn, error)
}
