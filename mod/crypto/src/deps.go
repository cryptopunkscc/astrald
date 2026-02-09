package crypto

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/core"
)

func (mod *Module) LoadDependencies(*astral.Context) (err error) {
	err = core.Inject(mod.node, &mod.Deps)
	if err != nil {
		return err
	}

	keyID, err := mod.Objects.Store(astral.NewContext(nil), mod.Objects.System(), mod.nodeKey)
	if err != nil {
		mod.log.Error("failed to store node key: %v", err)
	} else {
		mod.log.Log("node key id: %v", keyID)
	}

	mod.indexPrivateKey(mod.nodeKey)

	return err
}
