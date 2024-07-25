package profile

import (
	"context"
	"encoding/json"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/id"
	"github.com/cryptopunkscc/astrald/mod/profile/proto"
)

type EventHandler struct {
	*Module
}

func (h *EventHandler) Run(ctx context.Context) error {
	<-ctx.Done()
	return nil
}

func (h *EventHandler) updateIdentityProfile(target id.Identity, serviceName string) error {
	h.log.Infov(2, "updating profile for %s", target)

	conn, err := astral.Route(h.ctx, h.node.Router(), astral.NewQuery(h.node.Identity(), target, serviceName))
	if err != nil {
		return err
	}
	defer conn.Close()

	var profile proto.Profile
	err = json.NewDecoder(conn).Decode(&profile)
	if err != nil {
		return err
	}

	for _, pep := range profile.Endpoints {
		ep, err := h.exonet.Parse(pep.Network, pep.Address)
		if err != nil {
			continue
		}

		_ = h.nodes.AddEndpoint(target, ep)
	}

	h.log.Info("%s profile updated.", target)

	return nil
}
