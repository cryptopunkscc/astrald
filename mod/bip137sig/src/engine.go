package src

import (
	"encoding/base64"
	"fmt"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/crypto"
	"github.com/cryptopunkscc/astrald/mod/secp256k1"
	"github.com/cryptopunkscc/bip-0137/verify"
	secp "github.com/decred/dcrd/dcrec/secp256k1/v4"
)

type Engine struct {
	mod *Module
	crypto.NilEngine
}

func (e Engine) MessageSigner(key *crypto.PublicKey, scheme string) (crypto.MessageSigner, error) {
	switch {
	case scheme != crypto.SchemeBIP137:
		return nil, crypto.ErrUnsupportedScheme
	case !isBip137KeyType(key.Type):
		return nil, crypto.ErrUnsupportedKeyType
	}

	privateKey, err := e.mod.Crypto.PrivateKey(astral.NewContext(nil), key)
	if err != nil {
		return nil, fmt.Errorf("failed to get private key: %w", err)
	}

	compressed, err := isCompressedPublicKey(key.Key)
	if err != nil {
		return nil, err
	}

	return NewMessageSignerBip137(privateKey, compressed), nil
}

func (e Engine) VerifyMessageSignature(key *crypto.PublicKey, sig *crypto.Signature, msg string) error {
	switch {
	case !isBip137KeyType(key.Type):
		return crypto.ErrUnsupportedKeyType
	case sig.Scheme != crypto.SchemeBIP137:
		return crypto.ErrUnsupportedScheme
	}

	publicKey, err := secp.ParsePubKey(key.Key)
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

func isBip137KeyType(keyType astral.String8) bool {
	switch string(keyType) {
	case secp256k1.KeyType:
		return true
	default:
		return false
	}
}

func isCompressedPublicKey(key []byte) (bool, error) {
	switch len(key) {
	case 33:
		return true, nil
	case 65:
		return false, nil
	default:
		return false, fmt.Errorf("invalid public key length: %d", len(key))
	}
}
