package astral

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/hub"
	"github.com/cryptopunkscc/astrald/lib/astral"
	"github.com/cryptopunkscc/astrald/node"
	"io"
)

var _ astral.Api = &astralApi{}

type astralApi struct {
	ctx  context.Context
	node *node.Node
}

func (a *astralApi) Resolve(name string) (identity id.Identity, err error) {
	if name == "localnode" {
		return a.node.Identity(), nil
	}

	if identity, err = id.ParsePublicKeyHex(name); err == nil {
		return
	}

	identity, err = a.node.Contacts.ResolveIdentity(name)
	return
}

type astralPort struct {
	port hub.Port
}

type astralRequest struct {
	query *hub.Query
}

func (a *astralApi) Register(name string) (p astral.Port, err error) {
	hp, err := a.node.Ports.RegisterContext(a.ctx, name)
	if err != nil {
		return
	}
	p = &astralPort{*hp}
	return
}

func (a *astralApi) Query(nodeID id.Identity, query string) (rwc io.ReadWriteCloser, err error) {
	return a.node.Query(a.ctx, nodeID, query)
}

func (p *astralPort) Next() <-chan astral.Request {
	c := make(chan astral.Request)
	go func() {
		defer close(c)
		for query := range p.port.Queries() {
			q := query
			c <- &astralRequest{q}
		}
	}()
	return c
}

func (p *astralPort) Close() error {
	return p.port.Close()
}

func (r *astralRequest) Caller() (identity id.Identity) {
	if !r.query.IsLocal() {
		identity = r.query.Link().RemoteIdentity()
	}
	return
}

func (r *astralRequest) Query() string {
	return r.query.Query()
}

func (r *astralRequest) Accept() (io.ReadWriteCloser, error) {
	return r.query.Accept()
}

func (r *astralRequest) Reject() error {
	return r.query.Reject()
}
