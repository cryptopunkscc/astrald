package store

import (
	"errors"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/proto/block"
	"io"
)

type Store interface {
	Open(id data.ID, flags uint32) (block.Block, error)
	Create(alloc uint64) (block.Block, string, error)
	Download(blockID data.ID, offset uint64, limit uint64) (io.ReadCloser, error)
}

// ErrNotFound - block was not found in the store
var ErrNotFound = errors.New("not found")

// ErrUnsupported - operation not supported by the store
var ErrUnsupported = errors.New("unsupported")

const (
	OpenRemote = 1 << iota
)
