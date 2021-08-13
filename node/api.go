package node

import (
	"github.com/cryptopunkscc/astrald/api"
	_id "github.com/cryptopunkscc/astrald/node/auth/id"
	"github.com/cryptopunkscc/astrald/node/hub"
	_link "github.com/cryptopunkscc/astrald/node/link"
	"github.com/cryptopunkscc/astrald/node/peer"
)

type API struct {
	network api.Network
}

func (api API) Network() api.Network {
	return api.network
}

func NewAPI(localIdentity _id.Identity, peers *peer.Peers, hub *hub.Hub, linker *Linker) *API {
	return &API{
		network: &networkAPI{
			localIdentity: localIdentity,
			Linker:        linker,
			Peers:         peers,
			Hub:           hub,
		},
	}
}

type networkAPI struct {
	localIdentity _id.Identity
	*peer.Peers
	*Linker
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

func (_api *networkAPI) link(remoteID *_id.ECIdentity) (*_link.Link, error) {
	peer, _ := _api.Peers.Peer(remoteID)

	if peer.Connected() {
		return peer.DefaultLink(), nil
	}

	link, err := _api.Linker.Link(remoteID)
	if err != nil {
		return nil, err
	}

	_api.Peers.AddLink(link)

	return link, nil
}

func (_api *networkAPI) Connect(nodeID api.Identity, query string) (api.Stream, error) {
	remoteID, err := _id.ParsePublicKeyHex(string(nodeID))
	if err != nil {
		return nil, err
	}

	link, err := _api.link(remoteID)
	if err != nil {
		return nil, err
	}

	// Connect to identity's port
	return link.Query(query)
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
