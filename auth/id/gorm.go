package id

import (
	"database/sql/driver"
	"errors"
)

func (id Identity) Value() (driver.Value, error) {
	return id.PublicKeyHex(), nil
}

func (id *Identity) Scan(src any) error {
	str, ok := src.(string)
	if !ok {
		return errors.New("typcase failed")
	}

	parsed, err := ParsePublicKeyHex(str)
	if err != nil {
		return err
	}

	*id = parsed

	return nil
}
