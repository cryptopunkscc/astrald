package tree

import (
	"io"

	"github.com/cryptopunkscc/astrald/astral"
)

type ErrNodeHasNoValue struct {
}

var _ astral.Object = &ErrNodeHasNoValue{}

func (err ErrNodeHasNoValue) ObjectType() string {
	return "mod.tree.err_node_has_no_value"
}

func (err ErrNodeHasNoValue) WriteTo(io.Writer) (int64, error) {
	return 0, nil
}

func (err ErrNodeHasNoValue) ReadFrom(io.Reader) (int64, error) {
	return 0, nil
}

func (err *ErrNodeHasNoValue) Error() string {
	return "node has no value"
}

func (err *ErrNodeHasNoValue) Is(other error) bool {
	_, ok := other.(*ErrNodeHasNoValue)
	return ok
}

func init() {
	_ = astral.Add(&ErrNodeHasNoValue{})
}
