package profile

import (
	"encoding/json"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/mod/profile/proto"
)

func (mod *Module) ReceiveObject(drop objects.Drop) error {
	if !drop.SenderID().IsEqual(mod.node.Identity()) {
		return nil
	}

	switch obj := (drop.Object()).(type) {
	case *nodes.EventLinked:
		go mod.updateIdentityProfile(obj.NodeID)
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
