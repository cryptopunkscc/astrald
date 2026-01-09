package user

import (
	"io"
	"time"

	"github.com/cryptopunkscc/astrald/astral"
)

type NodeContractRevocation struct {
	ContractID *astral.ObjectID
	ExpiresAt  astral.Time // Required to have as we could not posses contract anymore
	CreatedAt  astral.Time // purely informational
}

var _ astral.Object = &NodeContractRevocation{}

func (c NodeContractRevocation) ObjectType() string {
	return "mod.users.node_contract_revocation"
}

func (c NodeContractRevocation) WriteTo(w io.Writer) (n int64, err error) {
	return astral.Struct(c).WriteTo(w)
}

func (c *NodeContractRevocation) ReadFrom(r io.Reader) (n int64, err error) {
	return astral.Struct(c).ReadFrom(r)
}

func (c NodeContractRevocation) IsExpired() bool {
	return time.Now().After(c.ExpiresAt.Time())
}

func init() {
	astral.Add(&NodeContractRevocation{})
}
