package identity

import (
	"github.com/cryptopunkscc/astrald/components/uid"
)

type Context struct {
	ids uid.Identities
}

type Request struct {
	*Context
}
