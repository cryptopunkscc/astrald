package embedded

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/hub"
	"github.com/cryptopunkscc/astrald/lib/wrapper"
	"github.com/cryptopunkscc/astrald/node"
	"io"
)

var _ wrapper.Api = &Adapter{}

type Adapter struct {
	Ctx  context.Context
	Node *node.Node
}

func (a *Adapter) Resolve(name string) (identity id.Identity, err error) {
	if name == "localnode" {
		return a.Node.Identity(), nil
	}

	if identity, err = id.ParsePublicKeyHex(name); err == nil {
		return
	}

	identity, err = a.Node.Contacts.ResolveIdentity(name)
	return
}

type astralPort struct {
	Node *node.Node
	port hub.Port
}

type astralRequest struct {
	Node  *node.Node
	query *hub.Query
}

func (a *Adapter) Register(name string) (p wrapper.Port, err error) {
	hp, err := a.Node.Ports.RegisterContext(a.Ctx, name)
	if err != nil {
		return
	}
	p = &astralPort{a.Node, *hp}
	return
}

func (a *Adapter) Query(nodeID id.Identity, query string) (rwc io.ReadWriteCloser, err error) {
	return a.Node.Query(a.Ctx, nodeID, query)
}

func (p *astralPort) Next() <-chan wrapper.Request {
	c := make(chan wrapper.Request)
	go func() {
		defer close(c)
		for query := range p.port.Queries() {
			q := query
			c <- &astralRequest{p.Node, q}
		}
	}()
	return c
}

func (p *astralPort) Close() error {
	return p.port.Close()
}

func (r *astralRequest) Caller() id.Identity {
	if r.query.IsLocal() {
		return r.Node.Identity()
	}
	return r.query.Link().RemoteIdentity()
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
