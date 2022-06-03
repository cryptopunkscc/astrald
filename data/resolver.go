package data

import (
	"bytes"
	"crypto/sha256"
	"hash"
	"io"
)

type Resolver interface {
	io.Writer
	Resolve() ID
}

type sha256Resolver struct {
	hash hash.Hash
	size uint64
}

func NewResolver() Resolver {
	return &sha256Resolver{
		hash: sha256.New(),
		size: 0,
	}
}

func (r *sha256Resolver) Write(p []byte) (n int, err error) {
	n, err = r.hash.Write(p)
	r.size = r.size + uint64(n)
	return
}

func (r sha256Resolver) Resolve() (id ID) {
	id.Size = r.size
	h := r.hash.Sum(nil)
	copy(id.Hash[0:32], h[0:32])
	return
}

func Resolve(data []byte) ID {
	r := NewResolver()
	b := bytes.NewReader(data)
	_, _ = io.Copy(r, b)
	return r.Resolve()
}

func ResolveAll(reader io.Reader) (ID, error) {
	r := NewResolver()

	if _, err := io.Copy(r, reader); err != nil {
		return ID{}, err
	}

	return r.Resolve(), nil
}
