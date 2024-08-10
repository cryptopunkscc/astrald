package objects

import (
	"github.com/cryptopunkscc/astrald/astral"
	"io"
)

type Stream struct {
	mod Module
	io.ReadWriter
}

type ObjectReader interface {
	ReadObject(r io.Reader) (astral.Object, error)
}

func NewStream(mod Module, c io.ReadWriter) *Stream {
	return &Stream{
		mod:        mod,
		ReadWriter: c,
	}
}

func (s *Stream) ReadObject() (astral.Object, error) {
	return s.mod.ReadObject(s)
}

func (s *Stream) WriteObject(obj astral.Object) (err error) {
	return WriteObject(s, obj)
}

func WriteObject(w io.Writer, obj astral.Object) (err error) {
	_, err = astral.ObjectHeader(obj.ObjectType()).WriteTo(w)
	if err != nil {
		return
	}
	_, err = obj.WriteTo(w)
	return
}
