package shell

import (
	"bufio"
	"github.com/cryptopunkscc/astrald/astral"
	"io"
)

var _ astral.ObjectReader = &LineReader{}

// LineReader reads lines of text from an io.Reader as astral.String
type LineReader struct {
	s *bufio.Scanner
}

func NewLineReader(r io.Reader) *LineReader {
	return &LineReader{
		s: bufio.NewScanner(r),
	}
}

func (r LineReader) ReadObject() (astral.Object, int64, error) {
	if r.s.Scan() {
		line := astral.String(r.s.Text())
		return &line, int64(len(line)), nil
	}

	return nil, 0, r.s.Err()
}
