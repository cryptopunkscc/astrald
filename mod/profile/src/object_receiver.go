package profile

import (
	"encoding/json"
	"slices"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/mod/events"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/mod/profile/proto"
)

func (mod *Module) ReceiveObject(drop objects.Drop) error {
	if !drop.SenderID().IsEqual(mod.node.Identity()) {
		return nil
	}

	switch o := (drop.Object()).(type) {
	case *events.Event:
		switch e := o.Data.(type) {
		case *nodes.StreamCreatedEvent:
			if e.StreamCount == 1 && slices.ContainsFunc(mod.User.LocalSwarm(),
				e.RemoteIdentity.IsEqual) {
				go mod.updateIdentityProfile(e.RemoteIdentity)
				drop.Accept(false)
			}

		}
	}

	return drop.Accept(false)
}

func (mod *Module) updateIdentityProfile(target *astral.Identity) error {
	mod.log.Infov(2, "updating profile for %v", target)

	conn, err := query.Route(mod.ctx.IncludeZone(astral.ZoneNetwork), mod.node, astral.NewQuery(mod.node.Identity(), target, serviceName))
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
		ep, err := mod.Exonet.Parse(pep.Network, pep.Address)
		if err != nil {
			continue
		}

		_ = mod.Nodes.AddEndpoint(target, ep)
	}

	mod.log.Info("%v profile updated.", target)

	return nil
}
