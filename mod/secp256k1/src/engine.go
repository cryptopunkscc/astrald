package secp256k1

import (
	"crypto/ecdsa"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/crypto"
	modSecp256k1 "github.com/cryptopunkscc/astrald/mod/secp256k1"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

// Engine implements KeyDeriver, HashVerifier, and HashSignerFactory for secp256k1 keys.
type Engine struct{}

// --- KeyDeriver ---

func (e Engine) PublicKey(ctx *astral.Context, key *crypto.PrivateKey) (*crypto.PublicKey, error) {
	return modSecp256k1.PublicKey(key), nil
}

// --- HashVerifier ---

func (e Engine) VerifyHash(key *crypto.PublicKey, sig *crypto.Signature, hash []byte) error {
	publicKey, err := secp256k1.ParsePubKey(key.Key)
	if err != nil {
		return err
	}

	if ecdsa.VerifyASN1(publicKey.ToECDSA(), hash, sig.Data) {
		return nil
	}

	return crypto.ErrInvalidSignature
}

// --- HashSignerFactory ---

func (e Engine) NewHashSigner(ctx *astral.Context, key *crypto.PrivateKey, scheme string) (crypto.HashSigner, error) {
	return modSecp256k1.NewHashSignerASN1(key), nil
}
