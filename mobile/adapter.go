package astralandroid

import (
	"context"
	"github.com/cryptopunkscc/astrald/api"
)

type serviceAdapter struct {
	delegate Service
}

func (s serviceAdapter) Run(
	ctx context.Context,
	core api.Core,
) error {
	//done := func() { <-ctx.Done() }
	return s.delegate.Run(
		//done,
		networkAdapter{delegate: core.Network()},
	)
}

type networkAdapter struct{ delegate api.Network }

func (n networkAdapter) Connect(
	identity string,
	port string,
) (Stream, error) {
	return n.delegate.Connect(api.Identity(identity), port)
}

func (n networkAdapter) Identity() string {
	return string(n.delegate.Identity())
}

func (n networkAdapter) Register(name string) (PortHandler, error) {
	h, err := n.delegate.Register(name)
	if err != nil {
		return nil, err
	}
	return portHandlerAdapter{
		delegate: h,
	}, nil
}

type portHandlerAdapter struct{ delegate api.PortHandler }

func (p portHandlerAdapter) Next() ConnectionRequest {
	return connectionRequestAdapter{delegate: <-p.delegate.Requests()}
}

func (p portHandlerAdapter) Close() error {
	return p.delegate.Close()
}

type connectionRequestAdapter struct{ delegate api.ConnectionRequest }

func (c connectionRequestAdapter) Caller() string {
	return string(c.delegate.Caller())
}

func (c connectionRequestAdapter) Query() string {
	return c.delegate.Query()
}

func (c connectionRequestAdapter) Accept() Stream {
	return c.delegate.Accept()
}

func (c connectionRequestAdapter) Reject() {
	c.delegate.Reject()
}

