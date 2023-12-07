package storage

import (
	"github.com/cryptopunkscc/astrald/data"
	"io"
	"time"
)

type DataManager interface {
	Storer
	Reader
	Indexer
	AddStorer(Storer)
	RemoveStorer(Storer)
	AddReader(Reader)
	RemoveReader(Reader)
	AddIndexer(indexer Indexer)
	RemoveIndexer(indexer Indexer)
}

type Storer interface {
	Store(int) (DataWriter, error)
}

type ReadOpts struct {
	Offset    uint64
	NoVirtual bool
}

type Reader interface {
	Read(id data.ID, opts *ReadOpts) (io.ReadCloser, error)
}

type Indexer interface {
	IndexSince(time time.Time) []DataInfo
}

type DataWriter interface {
	io.Writer
	Commit() (data.ID, error)
	Discard()
}

type DataInfo struct {
	ID        data.ID
	IndexedAt time.Time
}
