package policy

import "github.com/cryptopunkscc/astrald/auth/id"

type API interface {
	AddAlwaysLinkedIdentity(identity id.Identity) error
	RemoveAlwaysLinkedIdentity(identity id.Identity) error
}
