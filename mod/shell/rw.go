package shell

import (
	"github.com/cryptopunkscc/astrald/astral"
	"io"
)

type Input interface {
	Read() (astral.Object, error)
}

type Output interface {
	Write(astral.Object) error
}

type Stream interface {
	Input
	Output
}

type Join struct {
	Input
	Output
}

func NewStream(r io.Reader, w io.Writer) Stream {
	return &Join{
		Input:  NewLineReader(r),
		Output: NewStringWriter(w),
	}
}
