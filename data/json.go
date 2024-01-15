package data

import (
	"bytes"
	"encoding/json"
	"fmt"
)

func (id ID) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("\"%s\"", id.String())), nil
}

func (id *ID) UnmarshalJSON(b []byte) error {
	var s string
	var jsonDec = json.NewDecoder(bytes.NewReader(b))

	var err = jsonDec.Decode(&s)
	if err != nil {
		return err
	}

	parsed, err := Parse(s)
	if err != nil {
		return err
	}

	*id = parsed

	return nil
}
