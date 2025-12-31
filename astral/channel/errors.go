package channel

import (
	"errors"

	"github.com/cryptopunkscc/astrald/astral"
)

var ErrCloseUnsupported = errors.New("transport doesn't support closing")

type ReaderError struct {
	err error
}

func NewReaderError(err error) *ReaderError {
	return &ReaderError{err: err}
}

var _ Reader = &ReaderError{}

func (r ReaderError) Read() (astral.Object, error) {
	return nil, r.err
}

type WriterError struct {
	err error
}

func NewWriterError(err error) *WriterError {
	return &WriterError{err: err}
}

func (w WriterError) Write(astral.Object) error {
	return w.err
}
