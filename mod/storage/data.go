package storage

import (
	"github.com/cryptopunkscc/astrald/data"
	"time"
)

type DataManager interface {
	Reader
	Store
	Index
	ReadAll(id data.ID, opts *ReadOpts) ([]byte, error)
	StoreBytes(bytes []byte) (data.ID, error)
	AddReader(name string, reader Reader) error
	AddStore(name string, store Store) error
	AddIndex(name string, index Index) error
	RemoveReader(name string) error
	RemoveStore(name string) error
	RemoveIndex(name string) error
}

type Reader interface {
	Read(id data.ID, opts *ReadOpts) (DataReader, error)
}

type Store interface {
	Store(int) (DataWriter, error)
}

type Index interface {
	IndexSince(since time.Time) []DataInfo
}

type DataWriter interface {
	Write(p []byte) (n int, err error)
	Commit() (data.ID, error)
	Discard() error
}

type DataReader interface {
	Read(p []byte) (n int, err error)
	Close() error
}

type ReadOpts struct {
	Offset    uint64
	NoVirtual bool
}

type DataInfo struct {
	ID        data.ID
	IndexedAt time.Time
}
