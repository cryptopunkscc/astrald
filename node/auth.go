package node

import "github.com/cryptopunkscc/astrald/auth/id"

type AuthEngine interface {
	Authorize(id id.Identity, action string, args ...any) bool
	Add(Authorizer) error
	Remove(Authorizer) error
}

type Authorizer interface {
	Authorize(id id.Identity, action string, args ...any) bool
}
