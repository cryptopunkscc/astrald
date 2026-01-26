package src

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/core/assets"
	"github.com/cryptopunkscc/astrald/lib/ops"
	"github.com/cryptopunkscc/astrald/mod/bip137sig"
	"github.com/cryptopunkscc/astrald/mod/crypto"
	"github.com/cryptopunkscc/astrald/mod/secp256k1"
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

func (mod *Module) DeriveKey(seed bip137sig.Seed, path string) (privateKey crypto.PrivateKey, err error) {
	derivationPath, err := bip137sig.ParseDerivationPath(path)
	if err != nil {
		return privateKey, err
	}

	key, err := bip137sig.MasterKeyFromSeed(seed)
	if err != nil {
		return privateKey, err
	}

	for _, idx := range derivationPath {
		key, err = key.Derive(idx)
		if err != nil {
			return privateKey, err
		}
	}

	ecpPrivateKey, err := key.ECPrivKey()
	if err != nil {
		return
	}

	return crypto.PrivateKey{
		Type: secp256k1.KeyType,
		Key:  ecpPrivateKey.Serialize(),
	}, nil
}
