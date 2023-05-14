package cslq

import (
	"errors"
	"fmt"
	"io"
	"reflect"
)

// OpInterface codes value using Marshaler or Unmarshaler interfaces.
type OpInterface struct{}

// Unmarshaler is an interface implemented by objects that can decode CSLQ representation of themselves.
type Unmarshaler interface {
	UnmarshalCSLQ(dec *Decoder) error
}

// Marshaler is an interface implemented by objects that can encode CSLQ representation of themselves.
type Marshaler interface {
	MarshalCSLQ(enc *Encoder) error
}

// Formatter returns its own CSLQ pattern for encoding/decoding operations.
// NOTE: If Formatter is a struct, the returned pattern should be enclosed in "{}". Marshaler/Unmarshaler takes
// priority if also satisfied.
type Formatter interface {
	FormatCSLQ() string
}

func (op OpInterface) Encode(w io.Writer, data *Fifo) error {
	v := data.Pop()

	if m, ok := v.(Marshaler); ok {
		return m.MarshalCSLQ(NewEncoder(w))
	}

	if m, ok := v.(Formatter); ok {
		var f = m.FormatCSLQ()
		if isStruct(v) {
			f = "{" + f + "}"
		}
		return Encode(w, m.FormatCSLQ(), v)
	}

	if f, err := ScanStructFormat(v); err == nil {
		return Encode(w, "{"+f+"}", v)
	}

	if err, ok := v.(*error); ok {
		var errStr = ""
		if err != nil {
			errStr = (*err).Error()
		}
		return Encode(w, "[q]c", errStr)
	}

	return errors.New("variable does not implement Marshaler interface")
}

func (op OpInterface) Decode(r io.Reader, data *Fifo) error {
	v := data.Pop()

	if u, ok := v.(Unmarshaler); ok {
		return u.UnmarshalCSLQ(NewDecoder(r))
	}

	if m, ok := v.(Formatter); ok {
		var f = m.FormatCSLQ()
		if isStruct(v) {
			f = "{" + f + "}"
		}
		return Decode(r, f, v)
	}

	if f, err := ScanStructFormat(v); err == nil {
		return Decode(r, "{"+f+"}", v)
	}

	if err, ok := v.(*error); ok {
		var errStr string
		if err := Decode(r, "[q]c", &errStr); err != nil {
			return err
		}
		if errStr != "" {
			*err = errors.New(errStr)
		}
		return nil
	}

	return errors.New("variable does not implement Unmarshaler interface")
}

func (op OpInterface) String() string {
	return "v"
}

func ScanStructFormat(v any) (string, error) {
	var f string
	var t = reflect.TypeOf(v)

	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		return f, errors.New("v is not a struct nor a pointer to a struct")
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}

		tag := field.Tag.Get(tagCSLQ)
		if tag == "" {
			return f, fmt.Errorf("field %s has no %s tag", field.Name, tagCSLQ)
		}
		if tag == tagSkip {
			continue
		}

		f = f + tag
	}

	return f, nil
}

func isStruct(v interface{}) bool {
	var t = reflect.TypeOf(v)

	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	return t.Kind() == reflect.Struct
}
