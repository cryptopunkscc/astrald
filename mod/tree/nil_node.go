package tree

import (
	"github.com/cryptopunkscc/astrald/astral"
)

// NilNode returns ErrUnsupported for all operations. Embed it in your node to
// avoid having to explicitly implement unsupported interface methods:
//
//	type MyNode struct {
//		tree.NilNode
//	}
type NilNode struct{}

var _ Node = &NilNode{}

func (NilNode) Get(ctx *astral.Context, follow bool) (<-chan astral.Object, error) {
	return nil, &ErrNodeHasNoValue{}
}

func (NilNode) Set(ctx *astral.Context, object astral.Object) error {
	return ErrUnsupported
}

func (NilNode) Delete(ctx *astral.Context) error {
	return ErrUnsupported
}

func (NilNode) Sub(ctx *astral.Context) (map[string]Node, error) {
	return nil, nil
}

func (NilNode) Create(ctx *astral.Context, name string) (Node, error) {
	return nil, ErrUnsupported
}
