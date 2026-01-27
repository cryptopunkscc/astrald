package bip137sig

import "github.com/cryptopunkscc/astrald/mod/crypto"

const ModuleName = "bip137sig"

type Module interface {
	GenerateSeed() (seed Seed, err error)
	DeriveKey(seed Seed, path string) (privateKey crypto.PrivateKey, err error)
}
