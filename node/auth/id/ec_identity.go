package id

import (
	"encoding/hex"
	"errors"
	"github.com/btcsuite/btcd/btcec"
)

// ECIdentity is an eliptic-curve-based identity
type ECIdentity struct {
	privateKey *btcec.PrivateKey
	publicKey  *btcec.PublicKey
}

func GenerateECIdentity() (*ECIdentity, error) {
	var err error
	id := &ECIdentity{}

	id.privateKey, err = btcec.NewPrivateKey(btcec.S256())
	if err != nil {
		return nil, err
	}

	return id, nil
}

func ECIdentityFromPublicKey(key *btcec.PublicKey) *ECIdentity {
	return &ECIdentity{publicKey: key}
}

func ECIdentityFromBytes(data []byte) (*ECIdentity, error) {
	priv, _ := btcec.PrivKeyFromBytes(btcec.S256(), data)
	if priv == nil {
		return nil, errors.New("parse error")
	}
	return &ECIdentity{
		privateKey: priv,
	}, nil
}

func (id *ECIdentity) PublicKey() *btcec.PublicKey {
	if id.privateKey != nil {
		return id.privateKey.PubKey()
	}
	return id.publicKey
}

func (id *ECIdentity) PrivateKey() *btcec.PrivateKey {
	return id.privateKey
}

func (id ECIdentity) String() string {
	return hex.EncodeToString(id.PublicKey().SerializeCompressed())
}

func ParsePublicKeyHex(hexKey string) (*ECIdentity, error) {
	pkData, err := hex.DecodeString(hexKey)
	if err != nil {
		return nil, err
	}

	return ParsePublicKey(pkData)
}

func ParsePublicKey(pkData []byte) (*ECIdentity, error) {
	key, err := btcec.ParsePubKey(pkData, btcec.S256())
	if err != nil {
		return nil, err
	}

	return &ECIdentity{
		publicKey: key,
	}, nil
}
