package id

import (
	"encoding/hex"
	"errors"
	"github.com/btcsuite/btcd/btcec/v2"
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

	id.privateKey, err = btcec.NewPrivateKey()
	if err != nil {
		return Identity{}, err
	}

	return id, nil
}

func PublicKey(key *btcec.PublicKey) Identity {
	return Identity{publicKey: key}
}

func ParsePrivateKey(data []byte) (Identity, error) {
	priv, _ := btcec.PrivKeyFromBytes(data)
	if priv == nil {
		return Identity{}, errors.New("parse error")
	}
	return Identity{
		privateKey: priv,
	}, nil
}

func ParsePublicKey(pkData []byte) (Identity, error) {
	// All zeroes are valid zero identity
	if isAllNull(pkData) {
		return Identity{}, nil
	}

	key, err := btcec.ParsePubKey(pkData)
	if err != nil {
		return Identity{}, err
	}

	return Identity{
		publicKey: key,
	}, nil
}

func isAllNull(data []byte) bool {
	for _, b := range data {
		if b != 0 {
			return false
		}
	}
	return true
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

// IsEqual checks if the public key is the same as the other identity's or if both are zero
func (id Identity) IsEqual(other Identity) bool {
	if id.IsZero() {
		return other.IsZero()
	}
	if other.IsZero() {
		return false
	}
	return id.PublicKey().IsEqual(other.PublicKey())
}

func (id Identity) IsZero() bool {
	return (id.privateKey == nil) && (id.publicKey == nil)
}

// String returns a string representation of this identtity
func (id Identity) String() string {
	return id.PublicKeyHex()
}

func (id Identity) Fingerprint() string {
	hex := id.PublicKeyHex()
	return hex[0:8] + ":" + hex[len(hex)-8:]
}
