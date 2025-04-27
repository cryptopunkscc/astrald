package object

import (
	"database/sql/driver"
	"errors"
)

func (id ID) Value() (driver.Value, error) {
	return id.String(), nil
}

func (id *ID) Scan(src any) error {
	str, ok := src.(string)
	if !ok {
		return errors.New("typcase failed")
	}

	parsed, err := ParseID(str)
	if err != nil {
		return err
	}

	*id = *parsed

	return nil
}
