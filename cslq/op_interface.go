package cslq

import (
	"errors"
	"io"
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
		return Encode(w, "{"+m.FormatCSLQ()+"}", v)
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
		return Decode(r, "{"+m.FormatCSLQ()+"}", v)
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
