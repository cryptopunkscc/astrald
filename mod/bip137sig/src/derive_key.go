package src

import (
	corebip "github.com/cryptopunkscc/astrald/mod/bip137sig"
	"github.com/cryptopunkscc/astrald/mod/crypto"
	"github.com/cryptopunkscc/astrald/mod/secp256k1"
)

func (mod *Module) DeriveKey(seed corebip.Seed, path string) (privateKey crypto.PrivateKey, err error) {
	derivationPath, err := corebip.ParseDerivationPath(path)
	if err != nil {
		return privateKey, err
	}

	key, err := corebip.MasterKeyFromSeed(seed)
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
		Type: secp256k1.KeyType, // NOTE: is this secp256k1?
		Key:  ecpPrivateKey.Serialize(),
	}, nil
}
