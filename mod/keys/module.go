package keys

import (
	"fmt"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/data"
)

const ModuleName = "keys"
const DBPrefix = "keys__"

type Module interface {
	CreateKey(alias string) (identity id.Identity, dataID data.ID, err error)
	LoadPrivateKey(dataID data.ID) (*PrivateKey, error)
	FindIdentity(hex string) (id.Identity, error)
	Sign(identity id.Identity, hash []byte) ([]byte, error)
}

const PrivateKeyDataType = "keys.private_key"
const KeyTypeIdentity = "ecdsa-secp256k1"

type PrivateKey struct {
	Type  string `cslq:"[c]c"`
	Bytes []byte `cslq:"[c]c"`
}

type KeyDesc struct {
	KeyType   string
	PublicKey id.Identity
}

func (k KeyDesc) Type() string {
	return "mod.keys.private_key"
}
func (k KeyDesc) String() string {
	return fmt.Sprintf("Private key of {{%s}}", k.PublicKey.PublicKeyHex())
}
