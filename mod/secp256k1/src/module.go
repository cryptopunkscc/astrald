package secp256k1

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/core/assets"
	"github.com/cryptopunkscc/astrald/lib/routing"
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
	router routing.OpRouter
	db     *DB
}

func (mod *Module) Run(ctx *astral.Context) error {
	<-ctx.Done()
	return nil
}

func (mod *Module) Router() astral.Router {
	return &mod.router
}

func (mod *Module) RegisterCryptoCapabilities(ctx *astral.Context, reg *crypto.Registry) {
	engine := &Engine{}

	reg.RegisterKeyDeriver(secp256k1.KeyType, engine)
	reg.RegisterHashVerifier(secp256k1.KeyType, crypto.SchemeASN1, engine)
	reg.RegisterHashSignerFactory(secp256k1.KeyType, crypto.SchemeASN1, engine)
}

func (mod *Module) String() string {
	return secp256k1.ModuleName
}
