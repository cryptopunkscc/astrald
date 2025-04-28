package keys

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/cryptopunkscc/astrald/astral"
	"io"
	"strings"
)

const KeyTypeIdentity = "ecdsa-secp256k1"

type PrivateKey struct {
	Type  astral.String8
	Bytes astral.Bytes8
}

// astral

var _ astral.Object = &PrivateKey{}

func (PrivateKey) ObjectType() string {
	return "keys.private_key"
}

func (key PrivateKey) WriteTo(w io.Writer) (n int64, err error) {
	return astral.Struct(key).WriteTo(w)
}

func (key *PrivateKey) ReadFrom(r io.Reader) (n int64, err error) {
	return astral.Struct(key).ReadFrom(r)
}

// json

func (key *PrivateKey) UnmarshalJSON(bytes []byte) error {
	type alias PrivateKey
	var a alias

	err := json.Unmarshal(bytes, &a)
	if err != nil {
		return err
	}

	*key = PrivateKey(a)
	return nil
}

func (key PrivateKey) MarshalJSON() ([]byte, error) {
	type alias PrivateKey
	return json.Marshal(alias(key))
}

// text

func (key PrivateKey) MarshalText() (text []byte, err error) {
	s := fmt.Sprintf("%s:%s", key.Type, base64.StdEncoding.EncodeToString(key.Bytes))
	return []byte(s), nil
}

func (key *PrivateKey) UnmarshalText(text []byte) error {
	parts := strings.SplitN(string(text), ":", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid format")
	}

	bytes, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return err
	}

	key.Type = astral.String8(parts[0])
	key.Bytes = bytes

	return nil
}

// init

func init() {
	astral.DefaultBlueprints.Add(&PrivateKey{})
}
