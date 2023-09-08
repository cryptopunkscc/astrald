package net

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"io"
)

var _ SecureWriteCloser = &SecurePipeWriter{}

type SecurePipeWriter struct {
	*SourceField
	io.WriteCloser
	identity id.Identity
}

// Identity returns the identity of the reader side of the pipe
func (w *SecurePipeWriter) Identity() id.Identity {
	return w.identity
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
