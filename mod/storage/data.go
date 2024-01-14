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
	StoreBytes(bytes []byte, opts *StoreOpts) (data.ID, error)
	AddReader(name string, reader Reader) error
	AddStore(name string, store Store) error
	AddIndex(name string, index Index) error
	RemoveReader(name string) error
	RemoveStore(name string) error
	RemoveIndex(name string) error
}

type Reader interface {
	Read(dataID data.ID, opts *ReadOpts) (DataReader, error)
}

type Store interface {
	Store(opts *StoreOpts) (DataWriter, error)
}

type Index interface {
	IndexSince(since time.Time) []IndexEntry
}

type DataWriter interface {
	Write(p []byte) (n int, err error)
	Commit() (data.ID, error)
	Discard() error
}

type DataReader interface {
	Read(p []byte) (n int, err error)
	Close() error
	Info() *ReaderInfo
}

type ReaderInfo struct {
	Name string
}

type ReadOpts struct {
	Offset    uint64
	NoVirtual bool
}

type StoreOpts struct {
	Alloc int
}

type IndexEntry struct {
	ID        data.ID
	IndexedAt time.Time
}
