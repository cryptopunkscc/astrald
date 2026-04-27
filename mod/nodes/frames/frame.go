package frames

import (
	"fmt"
	"io"
)

type Frame interface {
	io.ReaderFrom
	io.WriterTo
	fmt.Stringer
}
