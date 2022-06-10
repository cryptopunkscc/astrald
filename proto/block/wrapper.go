package block

import (
	"github.com/cryptopunkscc/astrald/data"
	"io"
)

var _ Block = &Wrapper{}

// Wrapper implements the Block interface with all methods (except End) returning ErrUnavailable. If the wrapped object
// implements any methods of the interface, they will be called instead.
type Wrapper struct {
	object interface{}
}

func Wrap(object interface{}) *Wrapper {
	return &Wrapper{object: object}
}

func (w Wrapper) Read(p []byte) (n int, err error) {
	if typed, ok := w.object.(io.Reader); ok {
		return typed.Read(p)
	}

	return 0, ErrUnavailable
}

func (w Wrapper) Write(p []byte) (n int, err error) {
	if typed, ok := w.object.(io.Writer); ok {
		return typed.Write(p)
	}

	return 0, ErrUnavailable
}

func (w Wrapper) Seek(offset int64, whence int) (int64, error) {
	if typed, ok := w.object.(io.Seeker); ok {
		return typed.Seek(offset, whence)
	}

	return 0, ErrUnavailable
}

func (w Wrapper) Finalize() (data.ID, error) {
	if typed, ok := w.object.(Finalizer); ok {
		return typed.Finalize()
	}

	return data.ID{}, ErrUnavailable
}

func (w Wrapper) Close() error {
	if typed, ok := w.object.(io.Closer); ok {
		return typed.Close()
	}
	return ErrUnavailable
}
