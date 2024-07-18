package gateway

import (
	"context"
	"encoding/json"
	"github.com/cryptopunkscc/astrald/id"
	"github.com/cryptopunkscc/astrald/mod/discovery"
	"github.com/cryptopunkscc/astrald/astral"
	"time"
)

const SubscribeServiceName = ".gateway.subscribe"
const SubscribeServiceType = "mod.gateway.subscribe"
const defaultSubscriptionDuration = 24 * time.Hour

type SubscribeService struct {
	*Module
}

type Subscription struct {
	Status    string
	ExpiresAt time.Time `json:"expires_at,omitempty"`
}

func (srv *SubscribeService) Run(ctx context.Context) error {
	var err = srv.AddRoute(SubscribeServiceName, srv)
	if err != nil {
		return err
	}
	defer srv.RemoveRoute(SubscribeServiceName)

	if srv.sdp != nil {
		srv.sdp.AddServiceDiscoverer(srv)
		defer srv.sdp.RemoveServiceDiscoverer(srv)
	}

	<-ctx.Done()
	return nil
}

func (srv *SubscribeService) RouteQuery(ctx context.Context, query astral.Query, caller astral.SecureWriteCloser, hints astral.Hints) (astral.SecureWriteCloser, error) {
	return astral.Accept(query, caller, func(conn astral.Conn) {
		defer conn.Close()

		s := &Subscription{
			Status:    "ok",
			ExpiresAt: time.Now().Add(defaultSubscriptionDuration),
		}

		json.NewEncoder(conn).Encode(s)
	})
}

func (srv *SubscribeService) DiscoverServices(ctx context.Context, caller id.Identity, origin string) ([]discovery.Service, error) {
	return []discovery.Service{
		{
			Identity: srv.node.Identity(),
			Name:     SubscribeServiceName,
			Type:     SubscribeServiceType,
		},
	}, nil
}
