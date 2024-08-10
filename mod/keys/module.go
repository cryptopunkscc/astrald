package keys

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/object"
)

const ModuleName = "keys"
const DBPrefix = "keys__"

type Module interface {
	CreateKey(alias string) (*astral.Identity, object.ID, error)
	LoadPrivateKey(object.ID) (*PrivateKey, error)
	FindIdentity(hex string) (*astral.Identity, error)
	Sign(identity *astral.Identity, hash []byte) ([]byte, error)
}
