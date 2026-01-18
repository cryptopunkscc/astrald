package secp256k1

import (
	"crypto/ecdsa"
	"fmt"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/crypto"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

type Engine struct {
	Crypto crypto.Module
}

func NewEngine(crypto crypto.Module) *Engine {
	return &Engine{Crypto: crypto}
}

func (e Engine) MessageSigner(key *crypto.PublicKey, scheme string) (crypto.MessageSigner, error) {
	return nil, crypto.ErrUnsupported
}

func (e Engine) VerifyMessageSignature(key *crypto.PublicKey, sig *crypto.Signature, msg string) error {
	return crypto.ErrUnsupported
}

func (e Engine) PublicKey(ctx *astral.Context, key *crypto.PrivateKey) (*crypto.PublicKey, error) {
	return PublicKeyFromPrivateKey(key), nil
}

func (e Engine) HashSigner(key *crypto.PublicKey, scheme string) (crypto.HashSigner, error) {
	switch {
	case key.Type != KeyType:
		return nil, crypto.ErrUnsupported
	case scheme != "asn1":
		return nil, crypto.ErrUnsupported
	}

	privateKey, err := e.Crypto.PrivateKey(astral.NewContext(nil), key)
	if err != nil {
		return nil, fmt.Errorf("failed to get private key: %w", err)
	}

	return NewHashSignerASN1(privateKey), nil
}

func (e Engine) VerifyHashSignature(key *crypto.PublicKey, sig *crypto.Signature, hash []byte) error {
	switch {
	case key.Type != KeyType:
		return crypto.ErrUnsupported
	case sig.Scheme != "asn1":
		return crypto.ErrUnsupported
	}

	publicKey, err := secp256k1.ParsePubKey(key.Key)
	if err != nil {
		return err
	}

	if ecdsa.VerifyASN1(publicKey.ToECDSA(), hash, sig.Data) {
		return nil
	}

	return crypto.ErrInvalidSignature
}
