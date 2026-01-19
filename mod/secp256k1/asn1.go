package secp256k1

import (
	"crypto/ecdsa"
	"crypto/rand"

	"github.com/cryptopunkscc/astrald/mod/crypto"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

func SignASN1(key *crypto.PrivateKey, hash []byte) (*crypto.Signature, error) {
	switch {
	case key.Type != KeyType:
		return nil, crypto.ErrUnsupportedKeyType
	}

	sig, err := ecdsa.SignASN1(
		rand.Reader,
		secp256k1.PrivKeyFromBytes(key.Key).ToECDSA(),
		hash,
	)
	if err != nil {
		return nil, err
	}

	return &crypto.Signature{
		Scheme: "asn1",
		Data:   sig,
	}, nil
}

func VerifyASN1(key *crypto.PublicKey, hash []byte, sig *crypto.Signature) error {
	switch {
	case key.Type != KeyType:
		return crypto.ErrUnsupportedKeyType
	case sig.Scheme != "asn1":
		return crypto.ErrUnsupportedScheme
	}

	// parse the key
	pkey, err := secp256k1.ParsePubKey(key.Key)
	if err != nil {
		return err
	}

	// verify sig
	if ecdsa.VerifyASN1(pkey.ToECDSA(), hash, sig.Data) {
		return nil
	}

	return crypto.ErrInvalidSignature
}
