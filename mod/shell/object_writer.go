package shell

import (
	"bytes"
	"github.com/cryptopunkscc/astrald/astral"
	"io"
)

var _ astral.ObjectWriter = &ObjectWriter{}

type ObjectWriter struct {
	w io.Writer
}

func NewObjectWriter(w io.Writer) *ObjectWriter {
	return &ObjectWriter{w: w}
}

func (w ObjectWriter) WriteObject(object astral.Object) (n int64, err error) {
	var buf = &bytes.Buffer{}
	_, err = astral.Write(buf, object, false)
	if err != nil {
		return
	}

	if buf.Len() == 0 {
		return 0, nil
	}

	return astral.Bytes32(buf.Bytes()).WriteTo(w.w)
}
