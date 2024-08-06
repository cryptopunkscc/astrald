package keys

import (
	"fmt"
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

type KeyDesc struct {
	KeyType   string
	PublicKey *astral.Identity
}

func (k KeyDesc) Type() string {
	return "mod.keys.private_key"
}
func (k KeyDesc) String() string {
	return fmt.Sprintf("Private key of {{%s}}", k.PublicKey.String())
}
