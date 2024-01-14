package user

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/node/router"
)

type Identity struct {
	identity id.Identity
	cert     []byte
	routes   *router.PrefixRouter
}

func (i *Identity) Identity() id.Identity {
	return i.identity
}

func (i *Identity) Cert() []byte {
	return i.cert
}

func (i *Identity) Routes() *router.PrefixRouter {
	return i.routes
}
