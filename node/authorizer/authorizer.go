package authorizer

import "github.com/cryptopunkscc/astrald/auth/id"

type AuthSet interface {
	Authorizer
	Add(Authorizer) error
	Remove(Authorizer) error
}

type Authorizer interface {
	Authorize(id id.Identity, action string, args ...any) bool
}
