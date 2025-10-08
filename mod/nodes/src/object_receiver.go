package nodes

import (
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"time"
)

func (mod *Module) ReceiveObject(drop objects.Drop) error {
	info, ok := drop.Object().(*nodes.ServiceTTL)
	if !ok {
		return nil
	}

	drop.Accept(false)

	err := mod.db.SaveService(
		info.ProviderID,
		string(info.Name),
		int(info.Priority),
		time.Now().Add(time.Duration(info.TTL)),
	)

	return err
}
