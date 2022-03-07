package cslq

import "io"

// Op describes the basic interface of an endec op
type Op interface {
	Encode(w io.Writer, v *Fifo) error
	Decode(r io.Reader, v *Fifo) error
	String() string
}
