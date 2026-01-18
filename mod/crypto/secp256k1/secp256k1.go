package secp256k1

import (
	"crypto/ecdsa"
	"crypto/rand"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/crypto"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

const KeyType = "secp256k1"

// NewKey generates a new private key
func NewKey() *crypto.PrivateKey {
	pkey, err := secp256k1.GeneratePrivateKey()
	if err != nil {
		return nil
	}

	return &crypto.PrivateKey{
		Type: KeyType,
		Key:  pkey.Serialize(),
	}
}

func WrapKey(key *secp256k1.PrivateKey) *crypto.PrivateKey {
	return &crypto.PrivateKey{
		Type: KeyType,
		Key:  key.Serialize(),
	}
}

// FromIdentity returns Identity's public key
func FromIdentity(identity *astral.Identity) *crypto.PublicKey {
	return &crypto.PublicKey{
		Type: KeyType,
		Key:  identity.PublicKey().SerializeCompressed(),
	}
}

func IdentityFromPublicKey(key *crypto.PublicKey) *astral.Identity {
	pubKey, err := secp256k1.ParsePubKey(key.Key)
	if err != nil {
		return nil

	}

	return astral.IdentityFromPubKey(pubKey)
}

func PublicKeyFromPrivateKey(key *crypto.PrivateKey) *crypto.PublicKey {
	if key.Type != KeyType {
		return nil
	}

	return &crypto.PublicKey{
		Type: KeyType,
		Key:  secp256k1.PrivKeyFromBytes(key.Key).PubKey().SerializeCompressed(),
	}
}

func PublicKeyFromIdentity(identity *astral.Identity) *crypto.PublicKey {
	return &crypto.PublicKey{
		Type: KeyType,
		Key:  identity.PublicKey().SerializeCompressed(),
	}
}

func SignASN1(key *crypto.PrivateKey, hash []byte) ([]byte, error) {
	switch {
	case key.Type != KeyType:
		return nil, crypto.ErrUnsupportedKeyType
	}

	return ecdsa.SignASN1(
		rand.Reader,
		secp256k1.PrivKeyFromBytes(key.Key).ToECDSA(),
		hash,
	)
}

func VerifyASN1(key *crypto.PublicKey, hash []byte, sig []byte) error {
	switch {
	case key.Type != KeyType:
		return crypto.ErrUnsupportedKeyType
	}

	// parse the key
	pkey, err := secp256k1.ParsePubKey(key.Key)
	if err != nil {
		return err
	}

	// verify sig
	if ecdsa.VerifyASN1(pkey.ToECDSA(), hash, sig) {
		return nil
	}

	return crypto.ErrInvalidSignature
}

func SignObjectASN1(key *crypto.PrivateKey, object astral.Object) ([]byte, error) {
	objectID, err := astral.ResolveObjectID(object)
	if err != nil {
		return nil, err
	}

	return SignASN1(key, objectID.Hash[:])
}

func VerifyObjectASN1(key *crypto.PublicKey, object astral.Object, sig []byte) error {
	objectID, err := astral.ResolveObjectID(object)
	if err != nil {
		return err
	}

	return VerifyASN1(key, objectID.Hash[:], sig)
}
