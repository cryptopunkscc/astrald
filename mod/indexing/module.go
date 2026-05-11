package indexing

import "github.com/cryptopunkscc/astrald/astral"

const ModuleName = "indexing"
const DBPrefix = "indexing__"

const (
	MethodRegisterIndexer = "indexing.register_indexer"
	MethodSubscribe       = "indexing.subscribe"
	MethodRemoveIndex     = "indexing.remove_index"
)

type Module interface {
	RegisterIndexer(ctx *astral.Context, name string) (astral.Nonce, error)
	RemoveIndexer(ctx *astral.Context, nonce astral.Nonce) error
	UpdateIndexerState(ctx *astral.Context, nonce astral.Nonce, repoName string, version uint64) error
}
