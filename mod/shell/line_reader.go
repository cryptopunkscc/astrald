package shell

import (
	"bufio"
	"github.com/cryptopunkscc/astrald/astral"
	"io"
)

var _ Input = &LineReader{}

// LineReader reads lines of text from an io.Reader as astral.String
type LineReader struct {
	s *bufio.Scanner
}

func NewLineReader(r io.Reader) *LineReader {
	return &LineReader{
		s: bufio.NewScanner(r),
	}
}

func (r LineReader) Read() (astral.Object, error) {
	if r.s.Scan() {
		line := astral.String(r.s.Text())
		return &line, nil
	}

	return nil, r.s.Err()
}
