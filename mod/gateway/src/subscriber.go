package gateway

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/id"
	"github.com/cryptopunkscc/astrald/log"
	"time"
)

const minimumSubscriptionDuration = 15 * time.Minute
const subscribeRetryInterval = 60 * time.Second

type Subscriber struct {
	node    astral.Node
	log     *log.Logger
	gateway id.Identity
	cancel  context.CancelFunc
}

func (s *Subscriber) Gateway() id.Identity {
	return s.gateway
}

func NewSubscriber(gateway id.Identity, node astral.Node, log *log.Logger) *Subscriber {
	return &Subscriber{node: node, log: log, gateway: gateway}
}

func (s *Subscriber) Run(ctx context.Context) error {
	ctx, s.cancel = context.WithCancel(ctx)
	defer s.cancel()

	var expiresAt time.Time
	for {
		conn, err := astral.Route(ctx, s.node.Router(), astral.NewQuery(s.node.Identity(), s.gateway, SubscribeServiceName))
		if err != nil {
			select {
			case <-ctx.Done():
				return nil
			case <-time.After(subscribeRetryInterval):
			}
			continue
		}

		var info Subscription
		err = json.NewDecoder(conn).Decode(&info)
		conn.Close()

		if err != nil {
			select {
			case <-ctx.Done():
				return nil
			case <-time.After(subscribeRetryInterval):
			}
			continue
		}

		if info.Status != "ok" {
			return errors.New("subscription rejected")
		}

		expiresAt = info.ExpiresAt
		if time.Until(expiresAt) < minimumSubscriptionDuration {
			return errors.New("subscription too short")
		}

		s.log.Infov(2, "subscribed to %v until %v", s.gateway, expiresAt)

		select {
		case <-ctx.Done():
			return nil
		case <-time.After(time.Until(expiresAt) - time.Minute):
		}
	}
}

func (s *Subscriber) Cancel() {
	if s.cancel != nil {
		s.cancel()
	}
}
