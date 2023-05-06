package node

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node/config"
	"github.com/cryptopunkscc/astrald/node/contacts"
	"github.com/cryptopunkscc/astrald/node/event"
	"github.com/cryptopunkscc/astrald/node/infra"
	"github.com/cryptopunkscc/astrald/node/link"
	"github.com/cryptopunkscc/astrald/node/network"
	"github.com/cryptopunkscc/astrald/node/services"
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
	Infra() infra.Infra
	Network() Network
	Tracker() Tracker
	Contacts() Contacts
	Services() Services
}

type Network interface {
	Link(context.Context, id.Identity) (*link.Link, error)
	Events() *event.Queue
	AddLink(*link.Link) error
	Peers() *network.PeerSet
	Server() *network.Server
	AddSecureConn(conn net.SecureConn) error
}

type Tracker interface {
	Add(identity id.Identity, addr net.Endpoint, expiresAt time.Time) error
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
	Register(name string) (*services.Service, error)
	RegisterContext(ctx context.Context, name string) (*services.Service, error)
	Query(ctx context.Context, query string, link *link.Link) (*services.Conn, error)
}
