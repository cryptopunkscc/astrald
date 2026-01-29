package user

import (
	"fmt"
	"io"
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/crypto"
)

type NodeContract struct {
	UserID    *astral.Identity
	NodeID    *astral.Identity
	StartsAt  astral.Time
	ExpiresAt astral.Time
}

const DefaultNodeContractLength = 365 * 24 * time.Hour

var _ astral.Object = &NodeContract{}
var _ crypto.HashableContract = &NodeContract{}
var _ crypto.TextableContract = &NodeContract{}

func NewNodeContract(userID, nodeID *astral.Identity) *NodeContract {
	return &NodeContract{
		UserID:    userID,
		NodeID:    nodeID,
		StartsAt:  astral.Now(),
		ExpiresAt: astral.Time(time.Now().Add(DefaultNodeContractLength)),
	}
}

func (NodeContract) ObjectType() string {
	return "mod.user.node_contract"
}

func (c *NodeContract) IsExpired() bool {
	return time.Now().After(c.ExpiresAt.Time())
}

func (c *NodeContract) ActiveAt(t time.Time) bool {
	return t.After(c.StartsAt.Time()) && t.Before(c.ExpiresAt.Time())
}

func (c NodeContract) ContractText() string {
	return fmt.Sprintf(
		"Node:%s User:%s From:%s To:%s",
		c.NodeID.String(),
		c.UserID.String(),
		c.StartsAt,
		c.ExpiresAt,
	)
}

func (c NodeContract) ContractHash() []byte {
	id, err := astral.ResolveObjectID(&c)
	if err != nil {
		return nil
	}
	return id.Hash[:]
}

func (c NodeContract) WriteTo(w io.Writer) (n int64, err error) {
	return astral.Objectify(&c).WriteTo(w)
}

func (c *NodeContract) ReadFrom(r io.Reader) (n int64, err error) {
	return astral.Objectify(c).ReadFrom(r)
}

func (c NodeContract) MarshalJSON() ([]byte, error) {
	return astral.Objectify(&c).MarshalJSON()
}

func (c *NodeContract) UnmarshalJSON(bytes []byte) error {
	return astral.Objectify(c).UnmarshalJSON(bytes)
}

func init() {
	astral.Add(&NodeContract{})
}
