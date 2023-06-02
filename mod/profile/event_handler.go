package profile

import (
	"context"
	"encoding/json"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/mod/discovery"
	"github.com/cryptopunkscc/astrald/mod/profile/proto"
	"github.com/cryptopunkscc/astrald/node/event"
)

type EventHandler struct {
	*Module
}

func (h *EventHandler) Run(ctx context.Context) error {
	return event.Handle(ctx, h.node.Events(), h.handleServicesDiscovered)
}

func (h *EventHandler) handleServicesDiscovered(ctx context.Context, e discovery.EventServicesDiscovered) error {
	for _, srv := range e.Services {
		if e.Identity.IsEqual(h.node.Identity()) {
			continue
		}
		if srv.Type == serviceType {
			return h.updateIdentity(e.Identity, srv.Name)
		}
	}
	return nil
}

func (h *EventHandler) updateIdentity(identity id.Identity, serviceName string) error {
	h.log.Infov(2, "updating profile for %s", identity)
	conn, err := h.node.Query(context.Background(), identity, serviceName)
	if err != nil {
		return err
	}

	var profile proto.Profile
	err = json.NewDecoder(conn).Decode(&profile)
	if err != nil {
		return err
	}

	for _, pep := range profile.Endpoints {
		ep, err := h.node.Infra().Parse(pep.Network, pep.Address)
		if err != nil {
			continue
		}

		h.node.Tracker().Add(identity, ep, pep.ExpiresAt)
	}

	return nil
}
