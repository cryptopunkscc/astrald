package keys

import (
	"context"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/keys"
	"github.com/cryptopunkscc/astrald/mod/objects"
)

func (mod *Module) Prepare(ctx context.Context) error {
	// import node's private key
	var nodeID = mod.node.Identity()

	if _, err := mod.FindIdentity(nodeID.String()); err != nil {
		err = mod.importNodeIdentity()
		if err != nil {
			mod.log.Errorv(0, "error importing node identity: %v", err)
		}
	}

	return nil
}

func (mod *Module) importNodeIdentity() (err error) {
	pk := &keys.PrivateKey{
		Type:  keys.KeyTypeIdentity,
		Bytes: mod.node.Identity().PrivateKey().Serialize(),
	}

	ctx := astral.NewContext(nil).WithIdentity(mod.node.Identity())

	_, err = objects.Save(ctx, pk, mod.Objects.WriteDefault())

	return err
}
