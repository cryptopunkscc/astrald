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

func (a *astralApi) Resolve(name string) (string, error) {
	if identity, err := id.ParsePublicKeyHex(name); err == nil {
		return identity.String(), nil
	}

	if name == "localnode" {
		return a.node.Identity().String(), nil
	}

	if identity, err := a.node.Contacts.ResolveIdentity(name); err == nil {
		return identity.String(), nil
	} else {
		return "", err
	}
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

func (a *astralApi) Query(nodeID string, query string) (rwc io.ReadWriteCloser, err error) {
	var identity id.Identity
	if nodeID != "" && nodeID != "localnode" {
		if identity, err = id.ParsePublicKeyHex(nodeID); err != nil {
			return
		}
	}
	return a.node.Query(a.ctx, identity, query)
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
	return r.query.Accept()
}

func (r *astralRequest) Reject() {
	r.query.Reject()
}
