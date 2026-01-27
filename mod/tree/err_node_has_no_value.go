package tree

import (
	"github.com/cryptopunkscc/astrald/astral"
)

type ErrNoValue struct {
	astral.EmptyObject
}

var _ astral.Object = &ErrNoValue{}

func (err ErrNoValue) ObjectType() string {
	return "mod.tree.err_no_value"
}

func (err *ErrNoValue) Error() string {
	return "no value is set"
}

func (err *ErrNoValue) Is(other error) bool {
	_, ok := other.(*ErrNoValue)
	return ok
}

func init() {
	_ = astral.Add(&ErrNoValue{})
}
