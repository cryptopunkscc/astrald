package connect

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/infra"
	"github.com/cryptopunkscc/astrald/infra/gw"
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
	return gw.NewAddr(w.remote, w.local)
}

func (w wrapper) RemoteAddr() infra.Addr {
	return gw.NewAddr(w.remote, w.local)
}
