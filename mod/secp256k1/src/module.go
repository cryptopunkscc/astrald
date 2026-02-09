package secp256k1

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/core/assets"
	"github.com/cryptopunkscc/astrald/lib/ops"
	"github.com/cryptopunkscc/astrald/mod/crypto"
	"github.com/cryptopunkscc/astrald/mod/secp256k1"
)

type Deps struct {
	Crypto crypto.Module
}

type Module struct {
	Deps
	config Config
	node   astral.Node
	log    *log.Logger
	assets assets.Assets
	ops    ops.Set
	db     *DB
}

func (mod *Module) Run(ctx *astral.Context) error {
	<-ctx.Done()
	return nil
}

func (mod *Module) GetOpSet() *ops.Set {
	return &mod.ops
}

func (mod *Module) CryptoEngine() crypto.Engine {
	return &Engine{mod: mod}
}

func (mod *Module) String() string {
	return secp256k1.ModuleName
}
