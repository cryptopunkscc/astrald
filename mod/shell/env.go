package shell

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/term"
	"io"
)

var _ term.Printer = &Env{}
var _ io.ReadWriter = &Env{}

type Env struct {
	astral.ObjectReader
	astral.ObjectWriter
	r io.Reader
	w io.Writer
}

func (e *Env) Read(p []byte) (n int, err error) {
	return e.r.Read(p)
}

func (e *Env) Write(p []byte) (n int, err error) {
	return e.w.Write(p)
}

func NewTextEnv(r io.Reader, w io.Writer) *Env {
	return &Env{
		r:            r,
		w:            w,
		ObjectReader: NewLineReader(r),
		ObjectWriter: term.NewBasicPrinter(w, &term.DefaultTypeMap),
	}
}

func NewBinaryEnv(r io.Reader, w io.Writer) *Env {
	return &Env{
		r:            r,
		w:            w,
		ObjectReader: NewObjectReader(r),
		ObjectWriter: NewObjectWriter(w),
	}
}

func (e *Env) Print(objects ...astral.Object) (err error) {
	for _, object := range objects {
		_, err = e.WriteObject(object)
		if err != nil {
			return
		}
	}
	return
}

func (e *Env) Printf(f string, v ...interface{}) error {
	return term.Printf(e, f, v...)
}
