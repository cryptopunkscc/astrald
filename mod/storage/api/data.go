package storage

import (
	"github.com/cryptopunkscc/astrald/data"
	"io"
	"time"
)

type DataManager interface {
	Reader
	Store
	Index
	AddReader(name string, reader Reader) error
	AddStore(name string, store Store) error
	AddIndex(name string, index Index) error
	RemoveReader(name string) error
	RemoveStore(name string) error
	RemoveIndex(name string) error
}

type Reader interface {
	Read(id data.ID, opts *ReadOpts) (io.ReadCloser, error)
}

type Store interface {
	Store(int) (DataWriter, error)
}

type Index interface {
	IndexSince(since time.Time) []DataInfo
}

type DataWriter interface {
	io.Writer
	Commit() (data.ID, error)
	Discard()
}

type ReadOpts struct {
	Offset    uint64
	NoVirtual bool
}

type DataInfo struct {
	ID        data.ID
	IndexedAt time.Time
}
