package astral

import (
	"github.com/cryptopunkscc/astrald/id"
	"io"
)

type Secure interface {
	Identity() id.Identity
}

type SecureWriteCloser interface {
	Secure
	io.WriteCloser
}

type SecureReader interface {
	Secure
	io.Reader
}

var _ SecureWriteCloser = &SecurePipeWriter{}

type SecurePipeWriter struct {
	*SourceField
	io.WriteCloser
	identity id.Identity
}

func NewSecurePipeWriter(w io.WriteCloser, identity id.Identity) *SecurePipeWriter {
	return &SecurePipeWriter{
		SourceField: NewSourceField(nil),
		WriteCloser: w,
		identity:    identity,
	}
}

// Identity returns the identity of the reader side of the pipe
func (w *SecurePipeWriter) Identity() id.Identity {
	return w.identity
}

// Transport returns the underlying writer without an identity attached to it
func (w *SecurePipeWriter) Transport() any {
	return w.WriteCloser
}

var _ SecureReader = &SecurePipeReader{}
var _ SourceGetter = &SecurePipeReader{}

type SecurePipeReader struct {
	io.Reader
	w *SecurePipeWriter
}

// Identity returns the identity of the reader side of the pipe
func (r *SecurePipeReader) Identity() id.Identity {
	if r.w == nil {
		return id.Identity{}
	}
	return r.w.Identity()
}

func (r *SecurePipeReader) Source() any {
	return r.w
}

func SecurePipe(target id.Identity) (*SecurePipeReader, *SecurePipeWriter) {
	r, wc := io.Pipe()
	pw := &SecurePipeWriter{
		SourceField: NewSourceField(nil),
		WriteCloser: wc,
		identity:    target,
	}
	pr := &SecurePipeReader{
		Reader: r,
		w:      pw,
	}
	return pr, pw
}
