package astral

import (
	"crypto/sha256"
	"hash"
	"io"
)

// WriteResolver is an io.Writer that calculates the ID of the data written to it.
// Must be created with NewWriteResolver.
type WriteResolver struct {
	hash hash.Hash
	size uint64
	w    io.Writer
}

// ReadResolver is an io.Reader wrapper that resolves the ID of the data read from the underlying reader
type ReadResolver struct {
	r        io.Reader
	resolver *WriteResolver
}

// NewWriteResolver returns a new WriteResolver. If w is not nil, all writes will go to w first and the ID of the
// successfully written data will be calculated.
func NewWriteResolver(w io.Writer) *WriteResolver {
	return &WriteResolver{
		hash: sha256.New(),
		size: 0,
		w:    w,
	}
}

func (r *WriteResolver) Write(p []byte) (n int, err error) {
	if r.w == nil {
		n, err = r.hash.Write(p)
		r.size = r.size + uint64(n)
		return
	}

	n, err = r.w.Write(p)
	if n > 0 {
		r.hash.Write(p[:n])
	}

	return
}

func (r *WriteResolver) Resolve() (id *ObjectID) {
	id = &ObjectID{
		Size: r.size,
	}
	h := r.hash.Sum(nil)
	copy(id.Hash[0:32], h[0:32])
	return
}

// NewReadResolver returns a new instance of a Read
func NewReadResolver(r io.Reader) *ReadResolver {
	return &ReadResolver{r: r, resolver: NewWriteResolver(nil)}
}

func (r *ReadResolver) Read(p []byte) (n int, err error) {
	n, err = r.r.Read(p)

	if n > 0 {
		r.resolver.Write(p[:n])
	}

	return n, err
}

func (r *ReadResolver) Resolve() *ObjectID {
	return r.resolver.Resolve()
}

func Resolve(r io.Reader) (id *ObjectID, err error) {
	var p [8192]byte
	rr := NewReadResolver(r)
	for err == nil {
		_, err = rr.Read(p[:])
	}
	if err == io.EOF {
		return rr.Resolve(), nil
	}
	return
}
