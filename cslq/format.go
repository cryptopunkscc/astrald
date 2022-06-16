package cslq

import (
	"bytes"
	"io"
)

type Format []Op

func (f Format) Encode(w io.Writer, v ...interface{}) error {
	vars := NewFifo(v...)
	buf := &bytes.Buffer{}

	for _, op := range f {
		if err := op.Encode(buf, vars); err != nil {
			return err
		}
	}

	_, err := w.Write(buf.Bytes())
	return err
}

func (f Format) Decode(r io.Reader, v ...interface{}) error {
	data := NewFifo(v...)

	for _, op := range f {
		if err := op.Decode(r, data); err != nil {
			return err
		}
	}

	return nil
}

func (f Format) String() (s string) {
	for _, sub := range f {
		s = s + sub.String()
	}
	return
}
