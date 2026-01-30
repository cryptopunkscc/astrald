package user

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/crypto"

	"io"
)

var _ astral.Object = &SignedNodeContract{}

type SignedNodeContract struct {
	*NodeContract
	UserSig *crypto.Signature // asn1 or bip137
	NodeSig *crypto.Signature // always asn1
}

func (SignedNodeContract) ObjectType() string {
	return "mod.user.signed_node_contract"
}

func (c SignedNodeContract) WriteTo(w io.Writer) (n int64, err error) {
	return astral.Objectify(&c).WriteTo(w)
}

func (c *SignedNodeContract) ReadFrom(r io.Reader) (n int64, err error) {
	return astral.Objectify(c).ReadFrom(r)
}

func (c SignedNodeContract) MarshalJSON() ([]byte, error) {
	return astral.Objectify(&c).MarshalJSON()
}

func (c *SignedNodeContract) UnmarshalJSON(bytes []byte) error {
	return astral.Objectify(c).UnmarshalJSON(bytes)
}

func (c *SignedNodeContract) IsNil() bool { return c == nil || c.NodeContract == nil }

func init() {
	astral.Add(&SignedNodeContract{})
}
