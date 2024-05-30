package object

import (
	"bytes"
	"crypto/sha256"
	"hash"
	"io"
	"os"
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

func (r *WriteResolver) Resolve() (id ID) {
	id.Size = r.size
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

func (r *ReadResolver) Resolve() ID {
	return r.resolver.Resolve()
}

func Resolve(data []byte) ID {
	r := NewWriteResolver(nil)
	b := bytes.NewReader(data)
	_, _ = io.Copy(r, b)
	return r.Resolve()
}

func ResolveAll(reader io.Reader) (ID, error) {
	r := NewWriteResolver(nil)

	if _, err := io.Copy(r, reader); err != nil {
		return ID{}, err
	}

	return r.Resolve(), nil
}

func ResolveFile(path string) (ID, error) {
	file, err := os.Open(path)
	if err != nil {
		return ID{}, err
	}
	defer file.Close()

	fileID, err := ResolveAll(file)
	if err != nil {
		return ID{}, err
	}

	return fileID, nil
}
