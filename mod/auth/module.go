package auth

import "github.com/cryptopunkscc/astrald/id"

const ModuleName = "auth"

type Module interface {
	Authorize(id id.Identity, action string, args ...any) bool
	AddAuthorizer(Authorizer) error
}

type Authorizer interface {
	Authorize(id id.Identity, action string, args ...any) bool
}
