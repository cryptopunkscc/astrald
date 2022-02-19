package astral

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/hub"
	astral "github.com/cryptopunkscc/astrald/mod/apphost/client"
	"github.com/cryptopunkscc/astrald/node"
	"io"
)

var _ astral.Api = &astralApi{}

type astralApi struct {
	ctx  context.Context
	node *node.Node
}

type astralPort struct {
	port hub.Port
}

type astralRequest struct {
	query hub.Query
}

func (a *astralApi) Register(name string) (p astral.Port, err error) {
	hp, err := a.node.Ports.RegisterContext(a.ctx, name)
	if err != nil {
		return
	}
	p = &astralPort{*hp}
	return
}

func (a *astralApi) Query(nodeID string, query string) (rwc io.ReadWriteCloser, err error) {
	hex, err := id.ParsePublicKeyHex(nodeID)
	if err != nil {
		return
	}
	return a.node.Query(a.ctx, hex, query)
}

func (p *astralPort) Next() <-chan astral.Request {
	c := make(chan astral.Request)
	go func() {
		defer close(c)
		for query := range p.port.Queries() {
			c <- &astralRequest{*query}
		}
	}()
	return c
}

func (p *astralPort) Close() error {
	return p.port.Close()
}

func (r *astralRequest) Caller() string {
	if r.query.IsLocal() {
		return ""
	}
	return r.query.Link().RemoteIdentity().String()
}

func (r *astralRequest) Query() string {
	return r.query.Query()
}

func (r *astralRequest) Accept() (io.ReadWriteCloser, error) {
	return r.query.Accept(), nil
}

func (r *astralRequest) Reject() {
	r.query.Reject()
}
