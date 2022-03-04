package cslq

import (
	"errors"
	"io"
)

// Endec represents a struct which is simultanously Encoder and a Decoder.
type Endec struct {
	*Encoder
	*Decoder
	io.Closer
}

// NewEndec returns a new Endec over the provided io.ReadWriter.
func NewEndec(rw io.ReadWriter) *Endec {
	e := &Endec{
		Encoder: NewEncoder(rw),
		Decoder: NewDecoder(rw),
	}
	if closer, ok := rw.(io.Closer); ok {
		e.Closer = closer
	}
	return e
}

// Close closes the underlying transport. Returns an error if the transport doesn't satisfy
// io.Closer.
func (endec *Endec) Close() error {
	if endec.Closer == nil {
		return errors.New("unsupported")
	}

	return endec.Closer.Close()
}
