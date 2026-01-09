package astral

import (
	"errors"
	"fmt"
	"io"
)

// TypeEncoder is a function that encodes the object type to the writer.
type TypeEncoder func(io.Writer, string) (int64, error)

// TypeDecoder is a function that decodes the object type from the reader.
type TypeDecoder func(io.Reader) (string, int64, error)

// CanonicalTypeEncoder encodes the object type to the writer in its canonical form.
func CanonicalTypeEncoder(w io.Writer, t string) (n int64, err error) {
	if t == "" {
		return 0, errors.New("empty type")
	}

	var m int64

	m, err = Stamp{}.WriteTo(w)
	n = +m
	if err != nil {
		return
	}

	m, err = ObjectType(t).WriteTo(w)
	n = +m
	return
}

// CanonicalTypeDecoder decodes the object type from the reader in its canonical form.
func CanonicalTypeDecoder(r io.Reader) (t string, n int64, err error) {
	var m int64
	m, err = (&Stamp{}).ReadFrom(r)
	n += m
	if err != nil {
		return
	}

	m, err = (*ObjectType)(&t).ReadFrom(r)
	n += m
	return
}

// ShortTypeEncoder encodes the object type to the writer in its short form.
func ShortTypeEncoder(w io.Writer, t string) (int64, error) {
	if t == "" {
		return 0, errors.New("empty type")
	}
	
	return ObjectType(t).WriteTo(w)
}

// ShortTypeDecoder decodes the object type from the reader in its short form.
func ShortTypeDecoder(r io.Reader) (t string, n int64, err error) {
	n, err = (*ObjectType)(&t).ReadFrom(r)
	return
}

// IndexedTypeEncoder encodes the object type to the writer using an index.
func IndexedTypeEncoder(types []string) TypeEncoder {
	var rev = map[string]Uint8{}
	for idx, v := range types {
		rev[v] = Uint8(idx)
	}

	return func(w io.Writer, t string) (int64, error) {
		var code Uint8
		var ok bool

		code, ok = rev[t]
		if !ok {
			return 0, errors.New("invalid type")
		}

		return code.WriteTo(w)
	}
}

// IndexedTypeDecoder decodes the object type from the reader using an index.
func IndexedTypeDecoder(types []string) TypeDecoder {
	return func(r io.Reader) (t string, n int64, err error) {
		var code Uint8
		n, err = code.ReadFrom(r)
		if err != nil {
			return
		}

		if int(code) >= len(types) {
			err = fmt.Errorf("invalid type code %d", code)
			return
		}

		return types[code], n, nil
	}
}

// Canonical sets the encoder and decoder to the canonical form.
func Canonical() ConfigFunc {
	return func(config *endecConfig) {
		config.Encoder = CanonicalTypeEncoder
		config.Decoder = CanonicalTypeDecoder
	}
}
