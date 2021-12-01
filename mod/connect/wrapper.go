package connect

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/infra"
	"github.com/cryptopunkscc/astrald/infra/astral"
	"io"
)

var _ infra.Conn = &wrapper{}

type wrapper struct {
	io.ReadWriteCloser
	local    id.Identity
	remote   id.Identity
	outbound bool
}

func (w wrapper) Outbound() bool {
	return w.outbound
}

func (w wrapper) LocalAddr() infra.Addr {
	return astral.NewAddr(w.local, id.Identity{})
}

func (w wrapper) RemoteAddr() infra.Addr {
	return astral.NewAddr(w.remote, id.Identity{})
}
