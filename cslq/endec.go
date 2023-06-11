package cslq

import (
	"errors"
	"io"
)

// Encoder represents a CSLQ encoder.
type Encoder struct {
	w io.Writer
	c io.Closer
}

// Decoder represents a CSLQ decoder.
type Decoder struct {
	r io.Reader
}

// Endec represents both an Encoder and a Decoder.
type Endec struct {
	*Encoder
	*Decoder
}

// NewEncoder returns a new Encoder instance that writes to the provided io.Writer.
func NewEncoder(w io.Writer) (enc *Encoder) {
	enc = &Encoder{w: w}
	if c, ok := w.(io.Closer); ok {
		enc.c = c
	}
	return
}

// NewDecoder returns a new Decoder instance that reads from the provided io.Reader.
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{r: r}
}

// NewEndec returns a new Endec over the provided io.ReadWriter.
func NewEndec(rw io.ReadWriter) *Endec {
	return &Endec{
		Encoder: NewEncoder(rw),
		Decoder: NewDecoder(rw),
	}
}

// Encodef encodes vars according to the format.
func (enc *Encoder) Encodef(f string, v ...interface{}) error {
	if format, err := Compile(f); err != nil {
		return err
	} else {
		return format.Encode(enc.w, v...)
	}
}

// Encode encodes all arguments using the "v" format
func (enc *Encoder) Encode(v ...interface{}) error {
	for _, i := range v {
		if err := enc.Encodef("v", i); err != nil {
			return err
		}
	}
	return nil
}

// Decodef decodes vars using the format.
func (dec *Decoder) Decodef(f string, v ...interface{}) error {
	if format, err := Compile(f); err != nil {
		return err
	} else {
		return format.Decode(dec.r, v...)
	}
}

// Decode decodes a struct from the stream using the "v" format. Argument v must be a pointer to a struct.
func (dec *Decoder) Decode(v ...any) error {
	for _, i := range v {
		if err := dec.Decodef("v", i); err != nil {
			return err
		}
	}
	return nil
}

// Close closes the underlying transport if supported.
func (endec *Endec) Close() error {
	return endec.Encoder.Close()
}

// Close closes the underlying transport if supported.
func (enc *Encoder) Close() error {
	if enc.c == nil {
		return errors.New("unsupported")
	}
	return enc.c.Close()
}

// Encode encodes vars according to the pattern.
func Encode(w io.Writer, pattern string, v ...interface{}) error {
	return NewEncoder(w).Encodef(pattern, v...)
}

// Decode decodes vars using the pattern.
func Decode(r io.Reader, pattern string, v ...interface{}) error {
	return NewDecoder(r).Decodef(pattern, v...)
}
