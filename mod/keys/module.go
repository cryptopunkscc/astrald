package keys

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/object"
)

const ModuleName = "keys"
const DBPrefix = "keys__"

type Module interface {
	CreateKey(alias string) (*astral.Identity, object.ID, error)
	FindIdentity(hex string) (*astral.Identity, error)
	SignASN1(signer *astral.Identity, hash []byte) ([]byte, error)
	VerifyASN1(signer *astral.Identity, hash []byte, sig []byte) error
}
