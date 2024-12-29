package shell

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/term"
	"io"
)

var _ Output = &StringWriter{}

type StringWriter struct {
	w io.Writer
}

func NewStringWriter(w io.Writer) *StringWriter {
	return &StringWriter{w: w}
}

func (w StringWriter) Write(object astral.Object) (err error) {
	_, err = w.w.Write([]byte(term.Stringify(object)))

	return
}
