package policy

import (
	"github.com/cryptopunkscc/astrald/id"
)

const ModuleName = "policy"

type Module interface {
	AddAlwaysLinkedIdentity(identity id.Identity) error
	RemoveAlwaysLinkedIdentity(identity id.Identity) error
}
