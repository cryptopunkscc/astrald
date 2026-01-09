package user

import (
	"crypto/sha256"
	"io"

	"github.com/cryptopunkscc/astrald/astral"
)

type SignedNodeContractRevocation struct {
	*Revoker
	*NodeContractRevocation
	Attachments *astral.Bundle
}

var _ astral.Object = &SignedNodeContractRevocation{}

func (SignedNodeContractRevocation) ObjectType() string {
	return "mod.users.signed_node_contract_revocation"
}

func (c *SignedNodeContractRevocation) ReadFrom(r io.Reader) (n int64, err error) {
	return astral.Objectify(c).ReadFrom(r)
}

func (c SignedNodeContractRevocation) WriteTo(w io.Writer) (n int64, err error) {
	return astral.Objectify(&c).WriteTo(w)
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
	astral.Add(&SignedNodeContractRevocation{})
}
