package astralandroid

import (
	"context"
	"github.com/cryptopunkscc/astrald/api"
	"github.com/cryptopunkscc/astrald/bind/api"
	"github.com/cryptopunkscc/astrald/node"
)

func serviceRunner(service astralApi.Service) node.ServiceRunner {
	return func(ctx context.Context, core api.Core) error {
		return service.Run(networkAdapter{delegate: core.Network()})
	}
}

type networkAdapter struct{ delegate api.Network }

func (n networkAdapter) Connect(
	identity string,
	port string,
) (astralApi.Stream, error) {
	return n.delegate.Connect(api.Identity(identity), port)
}

func (n networkAdapter) Identity() string {
	return string(n.delegate.Identity())
}

func (n networkAdapter) Register(name string) (astralApi.PortHandler, error) {
	h, err := n.delegate.Register(name)
	if err != nil {
		return nil, err
	}
	return portHandlerAdapter{
		delegate: h,
	}, nil
}

type portHandlerAdapter struct{ delegate api.PortHandler }

func (p portHandlerAdapter) Next() astralApi.ConnectionRequest {
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

func (c connectionRequestAdapter) Accept() astralApi.Stream {
	return c.delegate.Accept()
}

func (c connectionRequestAdapter) Reject() {
	c.delegate.Reject()
}
