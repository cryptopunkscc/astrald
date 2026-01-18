package crypto

import (
	"encoding/hex"
	"encoding/json"
	"io"

	"github.com/cryptopunkscc/astrald/astral"
)

type Hash256 []byte

var _ astral.Object = &Hash256{}

func (Hash256) ObjectType() string {
	return "mod.crypto.hash256"
}

// binary

func (hash Hash256) WriteTo(w io.Writer) (int64, error) {
	return astral.Bytes8(hash).WriteTo(w)
}

func (hash *Hash256) ReadFrom(r io.Reader) (int64, error) {
	return (*astral.Bytes8)(hash).ReadFrom(r)
}

// json

func (hash Hash256) MarshalJSON() ([]byte, error) {
	return json.Marshal(hex.EncodeToString(hash))
}

func (hash *Hash256) UnmarshalJSON(bytes []byte) (err error) {
	var hashHex string
	err = json.Unmarshal(bytes, &hashHex)
	if err != nil {
		return
	}
	*hash, err = hex.DecodeString(hashHex)

	return
}

// text

func (hash Hash256) MarshalText() (text []byte, err error) {
	return []byte(hex.EncodeToString(hash)), nil
}

func (hash *Hash256) UnmarshalText(text []byte) (err error) {
	*hash, err = hex.DecodeString(string(text))
	return
}

// ...

func init() {
	_ = astral.Add(&Hash256{})
}
