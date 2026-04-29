package nodes

import (
	"io"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/mod/exonet"
)

type Link interface {
	channel.Receiver
	channel.Sender
	io.Closer
	CloseWithError(error) error
	Done() <-chan struct{}
	LocalIdentity() *astral.Identity
	RemoteIdentity() *astral.Identity
}

type NetworkLink interface {
	Link
	Network() string
	LocalEndpoint() exonet.Endpoint
	RemoteEndpoint() exonet.Endpoint
	Wake()
}

type QualityLink interface {
	Link
	Throughput() uint64
	IsHighPressure() bool
}
