package cslq

import "io"

// Endec represents a struct which is simultanously Encoder and a Decoder.
type Endec struct {
	*Encoder
	*Decoder
}

// NewEndec returns a new Endec over the provided io.ReadWriter
func NewEndec(rw io.ReadWriter) *Endec {
	return &Endec{
		Encoder: NewEncoder(rw),
		Decoder: NewDecoder(rw),
	}
}
