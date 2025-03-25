package user

import (
	"crypto/sha256"
	"github.com/cryptopunkscc/astrald/astral"
	"io"
)

var _ astral.Object = &SignedNodeContract{}

type SignedNodeContract struct {
	*NodeContract
	UserSig astral.Bytes8
	NodeSig astral.Bytes8
}

func (SignedNodeContract) ObjectType() string {
	return "mod.users.signed_node_contract"
}

func (c *SignedNodeContract) ReadFrom(r io.Reader) (n int64, err error) {
	return astral.Struct(c).ReadFrom(r)
}

func (c SignedNodeContract) WriteTo(w io.Writer) (n int64, err error) {
	return astral.Struct(c).WriteTo(w)
}

func (c *SignedNodeContract) Hash() []byte {
	var hash = sha256.New()
	_, err := c.NodeContract.WriteTo(hash)
	if err != nil {
		return nil
	}
	return hash.Sum(nil)
}
