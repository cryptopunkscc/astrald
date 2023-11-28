package gateway

import (
	"context"
	"encoding/json"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/mod/sdp"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node/modules"
	"time"
)

const SubscribeServiceName = "gateway.subscribe"
const defaultSubscriptionDuration = 24 * time.Hour

type SubscribeService struct {
	*Module
}

type Subscription struct {
	Status    string
	ExpiresAt time.Time `json:"expires_at,omitempty"`
}

func (srv *SubscribeService) Run(ctx context.Context) error {
	var err = srv.node.AddRoute(SubscribeServiceName, srv)
	if err != nil {
		return err
	}
	defer srv.node.RemoveRoute(SubscribeServiceName)

	if disco, err := modules.Find[*sdp.Module](srv.node.Modules()); err == nil {
		disco.AddSource(srv)
		defer disco.RemoveSource(srv)
	}

	<-ctx.Done()
	return nil
}

func (srv *SubscribeService) RouteQuery(ctx context.Context, query net.Query, caller net.SecureWriteCloser, hints net.Hints) (net.SecureWriteCloser, error) {
	return net.Accept(query, caller, func(conn net.SecureConn) {
		defer conn.Close()

		s := &Subscription{
			Status:    "ok",
			ExpiresAt: time.Now().Add(defaultSubscriptionDuration),
		}

		json.NewEncoder(conn).Encode(s)
	})
}

func (srv *SubscribeService) Discover(ctx context.Context, caller id.Identity, origin string) ([]sdp.ServiceEntry, error) {
	return []sdp.ServiceEntry{
		{
			Identity: srv.node.Identity(),
			Name:     SubscribeServiceName,
			Type:     "gateway.subscribe",
		},
	}, nil
}
