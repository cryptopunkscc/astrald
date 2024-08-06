package profile

import (
	"encoding/json"
	"errors"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/mod/profile/proto"
)

func (mod *Module) ReceiveObject(push *objects.Push) error {
	if !push.Source.IsEqual(mod.node.Identity()) {
		return errors.New("rejected")
	}

	switch obj := push.Object.(type) {
	case *nodes.EventLinked:
		go mod.updateIdentityProfile(obj.NodeID)
	}
	return errors.New("rejected")
}

func (mod *Module) updateIdentityProfile(target *astral.Identity) error {
	mod.log.Infov(2, "updating profile for %s", target)

	conn, err := astral.Route(mod.ctx, mod.node, astral.NewQuery(mod.node.Identity(), target, serviceName))
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

	mod.log.Info("%s profile updated.", target)

	return nil
}
