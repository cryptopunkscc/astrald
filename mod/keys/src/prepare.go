package keys

import (
	"context"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/mod/data"
	"github.com/cryptopunkscc/astrald/mod/keys"
	"github.com/cryptopunkscc/astrald/mod/storage"
)

func (mod *Module) Prepare(ctx context.Context) error {
	// import node's private key
	var nodeID = mod.node.Identity()

	if _, err := mod.FindIdentity(nodeID.PublicKeyHex()); err != nil {
		err = mod.importNodeIdentity()
		if err != nil {
			mod.log.Errorv(0, "error importing node identity: %v", err)
		}
	}

	return nil
}

func (mod *Module) importNodeIdentity() error {
	pk := keys.PrivateKey{
		Type:  keys.KeyTypeIdentity,
		Bytes: mod.node.Identity().PrivateKey().Serialize(),
	}

	w, err := mod.storage.Data().Store(&storage.StoreOpts{Alloc: 70})
	if err != nil {
		return err
	}

	err = cslq.Encode(w, "vv", data.ADC0Header(keys.PrivateKeyDataType), &pk)
	if err != nil {
		return err
	}

	_, err = w.Commit()

	return err
}
