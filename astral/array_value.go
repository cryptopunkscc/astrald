package astral

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
)

type arrayValue struct {
	reflect.Value
}

var _ Object = &arrayValue{}

// astral:blueprint-ignore
func (a arrayValue) ObjectType() string {
	return ""
}

func (a arrayValue) WriteTo(w io.Writer) (n int64, err error) {
	var o Object
	var m int64

	flagged := elemNeedsPresenceFlag(a.Type().Elem())
	for i := range a.Len() {
		o, err = objectify(a.Index(i))
		if err != nil {
			return
		}
		if flagged {
			if _, err = w.Write(presenceFlagOne); err != nil {
				return
			}
			n++
		}
		m, err = o.WriteTo(w)
		n += m
		if err != nil {
			return
		}
	}
	return
}

func (a arrayValue) ReadFrom(r io.Reader) (n int64, err error) {
	var o Object
	var m int64

	flagged := elemNeedsPresenceFlag(a.Type().Elem())
	for i := range a.Len() {
		o, err = objectify(a.Index(i))
		if err != nil {
			return
		}
		if flagged {
			m, err = consumePresenceFlag(r)
			n += m
			if err != nil {
				return
			}
		}
		m, err = o.ReadFrom(r)
		n += m
		if err != nil {
			return
		}
	}
	return
}

func (a arrayValue) MarshalJSON() (j []byte, err error) {
	// why: non-nil allocation so a zero-length array marshals to `[]` rather than `null`.
	arr := make([]json.RawMessage, 0, a.Len())
	var o Object
	var raw []byte

	for i := range a.Len() {
		o, err = objectify(a.Index(i))
		if err != nil {
			return
		}

		m, ok := o.(json.Marshaler)
		if !ok {
			return nil, errors.New("object does not implement json encoding")
		}

		raw, err = m.MarshalJSON()
		if err != nil {
			return
		}

		arr = append(arr, raw)
	}

	return json.Marshal(arr)
}

func (a arrayValue) UnmarshalJSON(bytes []byte) error {
	var arr []json.RawMessage

	err := json.Unmarshal(bytes, &arr)
	if err != nil {
		return err
	}
	// why: fixed-length array; oversize input previously panicked via reflect.Index out of
	// range. Match RuntimeArray.UnmarshalJSON's explicit error.
	if len(arr) != a.Len() {
		return fmt.Errorf("array_value: want %d elements, got %d", a.Len(), len(arr))
	}

	a.SetZero()

	for i, raw := range arr {
		o, err := objectify(a.Index(i))
		if err != nil {
			return err
		}

		m, ok := o.(json.Unmarshaler)
		if !ok {
			return errors.New("object does not implement json encoding")
		}

		err = m.UnmarshalJSON(raw)
		if err != nil {
			return err
		}
	}

	return nil
}
