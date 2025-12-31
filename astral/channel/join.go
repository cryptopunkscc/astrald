package channel

import "io"

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
