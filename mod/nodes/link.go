package nodes

import (
	"io"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/exonet"
)

type Link interface {
	astral.Router
	io.Closer

	ID() astral.Nonce
	Outbound() bool
	SetRouter(astral.Router)
	CloseWithError(error) error
	Done() <-chan struct{}
	LocalIdentity() *astral.Identity
	RemoteIdentity() *astral.Identity

	Network() string
	LocalEndpoint() exonet.Endpoint
	RemoteEndpoint() exonet.Endpoint
	Wake()

	Throughput() uint64
	IsHighPressure() bool
}
