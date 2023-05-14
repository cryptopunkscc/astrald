package node

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/node/contacts"
	"github.com/cryptopunkscc/astrald/node/event"
	"github.com/cryptopunkscc/astrald/node/infra"
	"github.com/cryptopunkscc/astrald/node/link"
	"github.com/cryptopunkscc/astrald/node/network"
	"github.com/cryptopunkscc/astrald/node/services"
	"github.com/cryptopunkscc/astrald/node/tracker"
)

type Node interface {
	Identity() id.Identity
	Alias() string
	SetAlias(alias string) error
	Query(ctx context.Context, remoteID id.Identity, query string) (link.BasicConn, error)
	Events() *event.Queue
	Infra() infra.Infra
	Network() network.Network
	Tracker() tracker.Tracker
	Contacts() contacts.Contacts
	Services() services.Services
}
