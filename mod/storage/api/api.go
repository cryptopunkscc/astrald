package storage

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/data"
	"io"
	"time"
)

type API interface {
	Reader
	AddReader(Reader)
	RemoveReader(Reader)
	CheckAccess(identity id.Identity, dataID data.ID) bool
	LocalFiles() LocalFiles
}

type LocalFiles interface {
	AddDir(ctx context.Context, path string) error
	RemoveDir(ctx context.Context, path string) error
	DataSince(time time.Time) []DataInfo
}

type DataInfo struct {
	ID        data.ID
	IndexedAt time.Time
}

type Reader interface {
	Read(id data.ID, offset int, length int) (io.ReadCloser, error)
}

var ErrNotFound = errors.New("not found")

type EventLocalFileAdded struct {
	Path string
	ID   data.ID
}

type EventLocalFileRemoved struct {
	Path string
	ID   data.ID
}

type EventLocalFileChanged struct {
	Path  string
	OldID data.ID
	NewID data.ID
}
