package astral

import (
	"bytes"
	"database/sql/driver"
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/btcsuite/btcd/btcec/v2"
	"io"
)

const anonymous = "anyone"

var Anyone = &Identity{}
var ErrInvalidKeyLength = errors.New("invalid key length")

// Identity is an eliptic-curve-based identity
type Identity struct {
	privateKey *btcec.PrivateKey
	publicKey  *btcec.PublicKey
}

// GenerateIdentity returns a new Identity
func GenerateIdentity() (*Identity, error) {
	var err error
	id := &Identity{}

	id.privateKey, err = btcec.NewPrivateKey()
	if err != nil {
		return nil, err
	}

	return id, nil
}

func IdentityFromString(s string) (*Identity, error) {
	if s == anonymous {
		return Anyone, nil
	}

	if len(s) != 66 {
		return nil, ErrInvalidKeyLength
	}

	pkBytes, err := hex.DecodeString(s)
	if err != nil {
		return nil, err
	}

	pk, err := btcec.ParsePubKey(pkBytes)
	if err != nil {
		return nil, err
	}

	return &Identity{publicKey: pk}, nil
}

func IdentityFromPubKey(key *btcec.PublicKey) *Identity {
	return &Identity{publicKey: key}
}

func IdentityFromPrivKeyBytes(privKey []byte) (*Identity, error) {
	priv, _ := btcec.PrivKeyFromBytes(privKey)
	if priv == nil {
		return nil, errors.New("parse error")
	}
	return &Identity{
		privateKey: priv,
	}, nil
}

// PublicKey returns identity's public key
func (id *Identity) PublicKey() *btcec.PublicKey {
	if id.privateKey != nil {
		return id.privateKey.PubKey()
	}
	return id.publicKey
}

// PrivateKey returns identity's private key
func (id *Identity) PrivateKey() *btcec.PrivateKey {
	return id.privateKey
}

// IsEqual checks if the public key is the same as the other identity's or if both are zero
func (id *Identity) IsEqual(other *Identity) bool {
	if id.IsZero() {
		return other.IsZero()
	}
	if other.IsZero() {
		return false
	}
	return id.PublicKey().IsEqual(other.PublicKey())
}

// IsZero returns true if identity is nil or has zero value
func (id *Identity) IsZero() bool {
	if id == nil {
		return true
	}
	return (id.privateKey == nil) && (id.publicKey == nil)
}

// String returns a string representation of this identity
func (id *Identity) String() string {
	if id.IsZero() {
		return anonymous
	}
	return hex.EncodeToString(id.PublicKey().SerializeCompressed())
}

func (id *Identity) Fingerprint() string {
	hex := id.String()
	return hex[0:8] + ":" + hex[len(hex)-8:]
}

func (id *Identity) WriteTo(w io.Writer) (n int64, err error) {
	if id.IsZero() {
		_, err = w.Write(make([]byte, btcec.PubKeyBytesLenCompressed))
	} else {
		_, err = w.Write(id.PublicKey().SerializeCompressed())
	}

	return
}

func (id *Identity) ReadFrom(r io.Reader) (n int64, err error) {
	var buf [btcec.PubKeyBytesLenCompressed]byte

	_, err = io.ReadFull(r, buf[:])
	if err != nil {
		return
	}

	id.publicKey, err = btcec.ParsePubKey(buf[:])
	if err != nil {
		return
	}

	return
}

func (id *Identity) Value() (driver.Value, error) {
	return id.String(), nil
}

func (id *Identity) Scan(src any) error {
	str, ok := src.(string)
	if !ok {
		return errors.New("typcast failed")
	}

	n, err := IdentityFromString(str)
	if err != nil {
		return err
	}

	*id = *n

	return nil
}

func (id *Identity) UnmarshalJSON(b []byte) error {
	var s string
	var jsonDec = json.NewDecoder(bytes.NewReader(b))

	var err = jsonDec.Decode(&s)
	if err != nil {
		return err
	}

	if s == anonymous {
		*id = *Anyone
		return nil
	}

	n, err := IdentityFromString(s)
	if err != nil {
		return err
	}

	*id = *n

	return nil
}

func (id *Identity) MarshalJSON() ([]byte, error) {
	if id.IsZero() {
		return []byte("\"" + anonymous + "\""), nil
	}
	return []byte("\"" + id.String() + "\""), nil
}

func (id *Identity) ObjectType() string {
	return "astral.identity.secp256k1"
}
