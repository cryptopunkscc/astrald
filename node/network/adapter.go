package network

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	iastral "github.com/cryptopunkscc/astrald/infra/astral"
	"io"
)

var _ iastral.Node = &Adapter{}

type Adapter struct {
	n *Network
}

func NewAdapter(n *Network) *Adapter {
	return &Adapter{n: n}
}

func (a Adapter) Identity() id.Identity {
	return a.n.localID
}

func (a Adapter) Query(ctx context.Context, remoteID id.Identity, query string) (io.ReadWriteCloser, error) {
	return a.n.Query(ctx, remoteID, query)
}
