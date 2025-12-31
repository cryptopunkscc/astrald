package channel

import "io"

// Joined combines an io.Reader and an io.Writer into a single io.ReadWriteCloser. Its Close method tries to typecast
// the Writer, then the Reader to an io.Closer. Returns ErrCloseUnsupported if neither implements io.Closer.
type Joined struct {
	io.Reader
	io.Writer
}

func Join(r io.Reader, w io.Writer) *Joined {
	return &Joined{r, w}
}

func (j Joined) Close() error {
	if c, ok := j.Writer.(io.Closer); ok {
		return c.Close()
	}
	if c, ok := j.Reader.(io.Closer); ok {
		return c.Close()
	}
	return ErrCloseUnsupported
}
