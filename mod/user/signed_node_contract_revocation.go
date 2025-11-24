package user

import (
	"crypto/sha256"
	"io"

	"github.com/cryptopunkscc/astrald/astral"
)

type SignedNodeContractRevocation struct {
	*NodeContractRevocation
	UserSig astral.Bytes8
}

var _ astral.Object = &SignedNodeContractRevocation{}

func (SignedNodeContractRevocation) ObjectType() string {
	return "mod.users.signed_node_contract_revocation"
}

func (c *SignedNodeContractRevocation) ReadFrom(r io.Reader) (n int64, err error) {
	return astral.Struct(c).ReadFrom(r)
}

func (c SignedNodeContractRevocation) WriteTo(w io.Writer) (n int64, err error) {
	return astral.Struct(c).WriteTo(w)
}

func (c *SignedNodeContractRevocation) Hash() []byte {
	var hash = sha256.New()
	_, err := c.NodeContractRevocation.WriteTo(hash)
	if err != nil {
		return nil
	}
	return hash.Sum(nil)
}

func init() {
	astral.DefaultBlueprints.Add(&SignedNodeContractRevocation{})
}
