package nodes

import (
	"context"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/id"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/tasks"
	"time"
)

const ModuleName = "nodes"
const DBPrefix = "nodes__"

type Module interface {
	exonet.Resolver
	Link(context.Context, id.Identity, LinkOpts) (Link, error)
	AcceptLink(ctx context.Context, conn exonet.Conn) (Link, error)
	InitLink(ctx context.Context, conn exonet.Conn, remoteID id.Identity) (Link, error)

	ParseInfo(s string) (*NodeInfo, error)

	AddEndpoint(id.Identity, ...exonet.Endpoint) error
	RemoveEndpoint(id.Identity, ...exonet.Endpoint) error

	Endpoints(id.Identity) []exonet.Endpoint

	Peers() []id.Identity
}

type NodeInfo struct {
	Identity  id.Identity
	Alias     string
	Endpoints []exonet.Endpoint
}

type Info struct {
	Linked      bool
	LastLinked  time.Time
	FirstLinked time.Time
}

type Endpoint struct {
	Network string
	Address string
}

type Desc struct {
	Endpoints []Endpoint
}

func (Desc) Type() string {
	return "node"
}

type LinkOpts struct {
	Endpoints []exonet.Endpoint
	Workers   int
}

// Link is an encrypted communication channel between two identities that can route queries
type Link interface {
	tasks.Runner
	astral.Router
	SetLocalRouter(astral.Router)
	LocalIdentity() id.Identity
	RemoteIdentity() id.Identity
	Transport() astral.Conn
	Close() error
	Done() <-chan struct{}
}
