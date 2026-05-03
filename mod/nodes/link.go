package nodes

import (
	"io"
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/exonet"
)

type Link interface {
	astral.Router
	io.Closer
	SetRouter(astral.Router)

	ID() astral.Nonce
	Outbound() bool
	CloseWithError(error) error
	Done() <-chan struct{}
	LocalIdentity() *astral.Identity
	RemoteIdentity() *astral.Identity

	Network() string
	LocalEndpoint() exonet.Endpoint
	RemoteEndpoint() exonet.Endpoint
	Wake()

	SetPressureDetector(detector LinkPressureDetector)
	Throughput() uint64
	IsHighPressure() bool
}

type LinkPressureDetector interface {
	OnBytes(n int, now time.Time)
	OnRTT(rtt time.Duration, now time.Time)
	IsHigh() bool
}
