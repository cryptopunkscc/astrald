package store

import (
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/proto/block"
)

type Store interface {
	Open(id data.ID, flags uint32) (block.Block, error)
	Create(alloc uint64) (block.Block, string, error)
}
