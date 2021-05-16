package node

import (
	"context"
	"github.com/cryptopunkscc/astrald/api"
	"github.com/cryptopunkscc/astrald/node/auth"
	"github.com/cryptopunkscc/astrald/node/hub"
	"github.com/cryptopunkscc/astrald/node/link"
	"github.com/cryptopunkscc/astrald/node/net"
	"github.com/cryptopunkscc/astrald/node/router"
)

type API struct {
	network api.Network
}

func (api API) Network() api.Network {
	return api.network
}

func NewAPI(localIdentity auth.Identity, router *router.Router, hub *hub.Hub) *API {
	return &API{
		network: &networkAPI{
			localIdentity: localIdentity,
			Router:        router,
			Hub:           hub,
		},
	}
}

type networkAPI struct {
	localIdentity auth.Identity
	*router.Router
	*hub.Hub
}

func (_api *networkAPI) Identity() api.Identity {
	return api.Identity(_api.localIdentity.String())
}

func (_api *networkAPI) Register(name string) (api.PortHandler, error) {
	port, err := _api.Hub.Register(name)
	if err != nil {
		return nil, err
	}

	return portAdapter{&port}, nil
}

func (_api *networkAPI) Connect(identity api.Identity, port string) (api.Stream, error) {
	var l *link.Link
	var err error
	ctx := context.Background()

	// No identity means connect to the local hub
	if (identity == "") || (string(identity) == _api.localIdentity.String()) {
		stream, err := _api.Hub.Connect(port, _api.localIdentity)
		if err != nil {
			return nil, err
		}

		return stream, nil
	}

	id, err := auth.ParsePublicKeyHex(string(identity))
	if err != nil {
		return nil, err
	}

	// Establish a link with the identity
	l, err = _api.Router.Connect(ctx, id)
	if err != nil {
		return nil, net.ErrHostUnreachable
	}

	// Connect to identity's port
	return l.Open(port)
}

type requestAdapter struct {
	*hub.Request
}

func (adapter requestAdapter) Accept() api.Stream {
	return adapter.Request.Accept()
}

func (adapter requestAdapter) Caller() api.Identity {
	return api.Identity(adapter.Request.Caller().String())
}

type portAdapter struct {
	*hub.Port
}

func (adapter portAdapter) Requests() <-chan api.ConnectionRequest {
	c := make(chan api.ConnectionRequest)
	go func() {
		for r := range adapter.Port.Requests() {
			c <- &requestAdapter{r}
		}
	}()

	return c
}
