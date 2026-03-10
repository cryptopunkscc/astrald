package gateway

import (
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/sig"
)

type binderConnPool struct {
	*Module

	conns sig.Set[exonet.Conn]
}

func newBinderConnPool(module *Module) *binderConnPool {
	return &binderConnPool{Module: module}
}

func (p *binderConnPool) add(conn exonet.Conn) {
	p.conns.Add(conn)
}

func (p *binderConnPool) take() (exonet.Conn, bool) {
	items := p.conns.Clone()
	if len(items) == 0 {
		return nil, false
	}
	conn := items[0]
	if err := p.conns.Remove(conn); err != nil {
		return nil, false
	}
	return conn, true
}
