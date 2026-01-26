package src

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/core/assets"
	"github.com/cryptopunkscc/astrald/lib/ops"
	"github.com/cryptopunkscc/astrald/mod/bip137sig"
	"github.com/cryptopunkscc/astrald/mod/crypto"
)

type Deps struct {
	Crypto crypto.Module
}

type Module struct {
	Deps
	node   astral.Node
	log    *log.Logger
	assets assets.Assets
	scope  ops.Set
}

var _ bip137sig.Module = &Module{}

func (mod *Module) Run(ctx *astral.Context) error {
	<-ctx.Done()
	return nil
}

func (mod *Module) GetOpSet() *ops.Set {
	return &mod.scope
}

func (mod *Module) String() string {
	return bip137sig.ModuleName
}
