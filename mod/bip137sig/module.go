package bip137sig

import "github.com/cryptopunkscc/astrald/mod/crypto"

const ModuleName = "bip137sig"

const (
	MethodDeriveKey  = "bip137sig.derive_key"
	MethodMnemonic   = "bip137sig.mnemonic"
	MethodNewEntropy = "bip137sig.new_entropy"
	MethodSeed       = "bip137sig.seed"
)

type Module interface {
	GenerateSeed() (seed Seed, err error)
	DeriveKey(seed Seed, path string) (privateKey crypto.PrivateKey, err error)
}
