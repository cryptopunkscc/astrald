package tree

import (
	"errors"

	"github.com/cryptopunkscc/astrald/astral"
)

var ErrNodeHasSubnodes = astral.NewError("node has subnodes")
var ErrUnsupported = astral.NewError("unsupported")
var ErrTypeMismatch = errors.New("binding type mismatch")
