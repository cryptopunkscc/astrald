package gateway

import (
	"encoding/json"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/query"
	"io"
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

func (srv *SubscribeService) Run(ctx *astral.Context) error {
	var err = srv.AddRoute(SubscribeServiceName, srv)
	if err != nil {
		return err
	}
	defer srv.RemoveRoute(SubscribeServiceName)

	<-ctx.Done()
	return nil
}

func (srv *SubscribeService) RouteQuery(ctx *astral.Context, q *astral.Query, w io.WriteCloser) (io.WriteCloser, error) {
	return query.Accept(q, w, func(conn astral.Conn) {
		defer conn.Close()

		s := &Subscription{
			Status:    "ok",
			ExpiresAt: time.Now().Add(defaultSubscriptionDuration),
		}

		json.NewEncoder(conn).Encode(s)
	})
}
