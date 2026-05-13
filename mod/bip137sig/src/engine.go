package src

import (
	"encoding/base64"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/crypto"
	"github.com/cryptopunkscc/astrald/mod/secp256k1"
	"github.com/cryptopunkscc/bip-0137/verify"
	dcrdSecp "github.com/decred/dcrd/dcrec/secp256k1/v4"
)

type Engine struct {
	mod *Module
}

// --- TextVerifier ---

func (e Engine) VerifyText(key *crypto.PublicKey, sig *crypto.Signature, msg string) error {
	publicKey, err := dcrdSecp.ParsePubKey(key.Key)
	if err != nil {
		return err
	}

	sigBase64 := base64.StdEncoding.EncodeToString(sig.Data)

	ok, err := verify.VerifyWithPubKey(publicKey, msg, sigBase64)
	if err != nil {
		return err
	}
	if !ok {
		return crypto.ErrInvalidSignature
	}

	return nil
}

// --- TextSignerFactory ---

func (e Engine) NewTextSigner(ctx *astral.Context, key *crypto.PrivateKey, scheme string) (crypto.TextSigner, error) {
	privKey, _ := btcec.PrivKeyFromBytes(key.Key)

	return &MessageSigner{
		key:        privKey,
		compressed: true,
	}, nil
}

// DerivePublicKey converts a private key to its public key equivalent.
func DerivePublicKey(key *crypto.PrivateKey) *crypto.PublicKey {
	return &crypto.PublicKey{
		Type: secp256k1.KeyType,
		Key:  dcrdSecp.PrivKeyFromBytes(key.Key).PubKey().SerializeCompressed(),
	}
}
