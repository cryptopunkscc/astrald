package term

import (
	"io"
)

type Nil struct{}

func (Nil) ObjectType() string {
	return "nil"
}

func (Nil) WriteTo(w io.Writer) (n int64, err error) {
	return 0, nil
}

func (Nil) ReadFrom(r io.Reader) (n int64, err error) {
	return 0, nil
}

func (Nil) String() string {
	return "nil"
}
