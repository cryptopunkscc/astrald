package channel

import (
	"encoding"
	"fmt"
	"io"

	"github.com/cryptopunkscc/astrald/astral"
)

type TextWriter struct {
	WithType bool
	w        io.Writer
}

var _ Writer = &TextWriter{}

func NewTextWriter(w io.Writer, withType bool) *TextWriter {
	return &TextWriter{w: w, WithType: withType}
}

func (t TextWriter) Write(obj astral.Object) error {
	// check if the object is a TextMarshaler
	m, ok := obj.(encoding.TextMarshaler)
	if !ok {
		return ErrTextUnsupported
	}

	// marshal the object into text
	text, err := m.MarshalText()
	if err != nil {
		return err
	}

	if t.WithType {
		_, err = fmt.Fprintf(t.w, "#[%s] %s\n", obj.ObjectType(), string(text))
	} else {
		_, err = fmt.Fprintf(t.w, "%s\n", string(text))
	}

	return err
}
