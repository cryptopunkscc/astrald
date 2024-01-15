package id

import (
	"bytes"
	"encoding/json"
)

func (id *Identity) UnmarshalJSON(b []byte) error {
	var s string
	var jsonDec = json.NewDecoder(bytes.NewReader(b))

	var err = jsonDec.Decode(&s)
	if err != nil {
		return err
	}

	if s == "anyone" {
		*id = Anyone
		return nil
	}

	parsed, err := ParsePublicKeyHex(s)
	if err != nil {
		return err
	}

	*id = parsed

	return nil
}

func (id Identity) MarshalJSON() ([]byte, error) {
	if id.IsZero() {
		return []byte("\"anyone\""), nil
	}
	return []byte("\"" + id.PublicKeyHex() + "\""), nil
}
