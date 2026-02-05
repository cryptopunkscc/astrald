package astral

import (
	"bytes"
	"database/sql/driver"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"strings"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

const anonymous = "anyone"

var Anyone = &Identity{}
var anyoneKey = strings.Repeat("00", 33)
var ErrInvalidKeyLength = errors.New("invalid key length")

// Identity is an eliptic-curve-based identity
type Identity struct {
	publicKey *secp256k1.PublicKey
}

func ParseIdentity(s string) (*Identity, error) {
	switch {
	case s == anyoneKey, s == anonymous:
		return Anyone, nil
	case len(s) != 66:
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

func IdentityFromPubKey(key *secp256k1.PublicKey) *Identity {
	return &Identity{publicKey: key}
}

// PublicKey returns identity's public key
func (id *Identity) PublicKey() *secp256k1.PublicKey {
	return id.publicKey
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
	return id == nil || id.publicKey == nil
}

// String returns a string representation of this identity
func (id *Identity) String() string {
	if id.IsZero() {
		return anyoneKey
	}
	return hex.EncodeToString(id.PublicKey().SerializeCompressed())
}

func (id *Identity) Fingerprint() string {
	hex := id.String()
	return hex[0:8] + ":" + hex[len(hex)-8:]
}

// astral

func (Identity) ObjectType() string {
	return "identity"
}

func (id Identity) WriteTo(w io.Writer) (n int64, err error) {
	var m int
	if id.IsZero() {
		m, err = w.Write(make([]byte, btcec.PubKeyBytesLenCompressed))
	} else {
		m, err = w.Write(id.PublicKey().SerializeCompressed())
	}
	n = int64(m)

	return
}

func (id *Identity) ReadFrom(r io.Reader) (n int64, err error) {
	var buf [btcec.PubKeyBytesLenCompressed]byte

	nn, err := io.ReadFull(r, buf[:])
	n = int64(nn)
	if err != nil {
		return
	}

	// if all bytes are null set zero value
	var allNull = true
	for i := 0; i < len(buf); i++ {
		if buf[i] != 0 {
			allNull = false
			break
		}
	}
	if allNull {
		id.publicKey = nil
		return
	}

	id.publicKey, err = btcec.ParsePubKey(buf[:])

	return
}

// json

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

	n, err := ParseIdentity(s)
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

// text

func (id *Identity) MarshalText() (text []byte, err error) {
	return []byte(id.String()), nil
}

func (id *Identity) UnmarshalText(text []byte) (err error) {
	i, err := ParseIdentity(string(text))
	if err != nil {
		return
	}
	*id = *i
	return
}

// sql

func (id *Identity) Value() (driver.Value, error) {
	return id.String(), nil
}

func (id *Identity) Scan(src any) error {
	str, ok := src.(string)
	if !ok {
		return errors.New("typcast failed")
	}

	n, err := ParseIdentity(str)
	if err != nil {
		return err
	}

	*id = *n

	return nil
}

func init() {
	_ = Add(&Identity{})
}
