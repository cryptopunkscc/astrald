package astralmobile

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/hub"
	astral "github.com/cryptopunkscc/astrald/mod/apphost/client"
	"github.com/cryptopunkscc/astrald/node"
	"io"
)

func newApiAdapter(ctx context.Context, n *node.Node) astral.Api {
	return &apiAdapter{ctx, n}
}

type apiAdapter struct {
	ctx  context.Context
	node *node.Node
}

type portAdapter struct {
	port *hub.Port
}

type requestAdapter struct {
	query *hub.Query
}

func (a *apiAdapter) Register(name string) (p astral.Port, err error) {
	hp, err := a.node.Ports.RegisterContext(a.ctx, name)
	if err != nil {
		return
	}
	p = &portAdapter{hp}
	return
}

func (a *apiAdapter) Query(nodeID string, query string) (rwc io.ReadWriteCloser, err error) {
	hex, err := id.ParsePublicKeyHex(nodeID)
	if err != nil {
		return
	}
	return a.node.Query(a.ctx, hex, query)
}

func (p *portAdapter) Next() <-chan astral.Request {
	c := make(chan astral.Request)
	go func() {
		defer close(c)
		for query := range p.port.Queries() {
			c <- &requestAdapter{query}
		}
	}()
	return c
}

func (p *portAdapter) Close() error {
	return p.port.Close()
}

func (r *requestAdapter) Caller() string {
	return r.query.Link().RemoteIdentity().String()
}

func (r *requestAdapter) Query() string {
	return r.query.Query()
}

func (r *requestAdapter) Accept() (io.ReadWriteCloser, error) {
	return r.query.Accept(), nil
}

func (r *requestAdapter) Reject() {
	r.query.Reject()
}
