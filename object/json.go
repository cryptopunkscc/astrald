package object

import (
	"bytes"
	"encoding/json"
	"fmt"
)

func (id ID) MarshalJSON() ([]byte, error) {
	if id.IsZero() {
		return []byte("\"\""), nil
	}

	return []byte(fmt.Sprintf("\"%s\"", id.String())), nil
}

func (id *ID) UnmarshalJSON(b []byte) error {
	var s string
	var jsonDec = json.NewDecoder(bytes.NewReader(b))

	var err = jsonDec.Decode(&s)
	if err != nil {
		return err
	}

	parsed, err := ParseID(s)
	if err != nil {
		return err
	}

	*id = parsed

	return nil
}
