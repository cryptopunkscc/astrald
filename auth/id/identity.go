package id

import (
	"encoding/hex"
	"errors"
	"github.com/btcsuite/btcd/btcec"
)

// Identity is an eliptic-curve-based identity
type Identity struct {
	privateKey *btcec.PrivateKey
	publicKey  *btcec.PublicKey
}

var ErrInvalidKeyLength = errors.New("invalid key length")

// GenerateIdentity returns a new Identity
func GenerateIdentity() (Identity, error) {
	var err error
	id := Identity{}

	id.privateKey, err = btcec.NewPrivateKey(btcec.S256())
	if err != nil {
		return Identity{}, err
	}

	return id, nil
}

func PublicKey(key *btcec.PublicKey) Identity {
	return Identity{publicKey: key}
}

func ParsePrivateKey(data []byte) (Identity, error) {
	priv, _ := btcec.PrivKeyFromBytes(btcec.S256(), data)
	if priv == nil {
		return Identity{}, errors.New("parse error")
	}
	return Identity{
		privateKey: priv,
	}, nil
}

func ParsePublicKey(pkData []byte) (Identity, error) {
	key, err := btcec.ParsePubKey(pkData, btcec.S256())
	if err != nil {
		return Identity{}, err
	}

	return Identity{
		publicKey: key,
	}, nil
}

func ParsePublicKeyHex(hexKey string) (Identity, error) {
	if len(hexKey) != 66 {
		return Identity{}, ErrInvalidKeyLength
	}

	pkData, err := hex.DecodeString(hexKey)
	if err != nil {
		return Identity{}, err
	}

	return ParsePublicKey(pkData)
}

func (id Identity) Public() Identity {
	return Identity{
		publicKey: id.PublicKey(),
	}
}

// PublicKey returns identity's public key
func (id Identity) PublicKey() *btcec.PublicKey {
	if id.privateKey != nil {
		return id.privateKey.PubKey()
	}
	return id.publicKey
}

// PublicKeyHex returns a serialized, compressed, hex-encoded public key
func (id Identity) PublicKeyHex() string {
	return hex.EncodeToString(id.PublicKey().SerializeCompressed())
}

// PrivateKey returns identity's private key
func (id Identity) PrivateKey() *btcec.PrivateKey {
	return id.privateKey
}

// IsEqual compares this Identity to another one by checking if their public keys are equal
func (id Identity) IsEqual(other Identity) bool {
	return id.PublicKey().IsEqual(other.PublicKey())
}

func (id Identity) IsZero() bool {
	return (id.privateKey == nil) && (id.publicKey == nil)
}

// String returns a string representation of this identtity
func (id Identity) String() string {
	return id.PublicKeyHex()
}
