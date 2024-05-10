package objects

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/object"
)

type Finder interface {
	Find(context.Context, object.ID, *net.Scope) []id.Identity
}
