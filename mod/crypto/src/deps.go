package crypto

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/mod/crypto"
)

func (mod *Module) LoadDependencies(ctx *astral.Context) (err error) {
	err = core.Inject(mod.node, &mod.Deps)
	if err != nil {
		return err
	}

	mod.reg = crypto.NewRegistry()

	// scan modules looking for crypto engines to auto-register capabilities
	_ = core.EachLoadedModule(mod.node, func(m core.Module) error {
		p, ok := m.(crypto.EngineProvider)
		if !ok {
			return nil
		}

		p.RegisterCryptoCapabilities(ctx, mod.reg)
		return nil
	})

	// store node key in system repo
	keyID, err := mod.Objects.Store(ctx, mod.Objects.System(), mod.nodeKey)
	if err != nil {
		mod.log.Error("failed to store node key: %v", err)
	} else {
		mod.log.Log("node key id: %v", keyID)
	}

	// index the node key
	mod.indexPrivateKey(mod.nodeKey)

	return err
}
