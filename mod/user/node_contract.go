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
var _ crypto.SignableTextObject = &NodeContract{}

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

func (c *NodeContract) SignableText() string {
	return fmt.Sprintf(
		"Node %s represents %s until %s",
		c.NodeID.String(),
		c.UserID.String(),
		c.ExpiresAt.Time().Format("2006-01-02 15:04:05"),
	)
}

func (c *NodeContract) SignableHash() []byte {
	id, err := astral.ResolveObjectID(c)
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
