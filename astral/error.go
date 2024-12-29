package astral

import (
	"io"
)

type Error interface {
	Object
	error
}

var _ Error = &ErrorMessage{}

type ErrorMessage struct {
	err string
}

func (b ErrorMessage) ObjectType() string {
	return "astral.errors.error_message"
}

func (b ErrorMessage) WriteTo(w io.Writer) (n int64, err error) {
	return String16(b.err).WriteTo(w)
}

func (b *ErrorMessage) ReadFrom(r io.Reader) (n int64, err error) {
	return (*String16)(&b.err).ReadFrom(r)
}

func (b ErrorMessage) Error() string {
	return b.err
}

func (b ErrorMessage) String() string {
	return b.err
}

func NewError(s string) Error {
	return &ErrorMessage{
		err: s,
	}
}
