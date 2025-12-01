package user

import (
	"io"
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/streams"
)

type NodeContract struct {
	UserID    *astral.Identity
	NodeID    *astral.Identity
	StartsAt  astral.Time
	ExpiresAt astral.Time
}

var _ astral.Object = &NodeContract{}

func (NodeContract) ObjectType() string {
	return "mod.users.node_contract"
}

func (c *NodeContract) IsExpired() bool {
	return time.Now().After(c.ExpiresAt.Time())
}

func (c NodeContract) IsActive() bool {
	now := time.Now()
	return now.After(c.StartsAt.Time()) && now.Before(c.ExpiresAt.Time())
}

func (c NodeContract) WriteTo(w io.Writer) (n int64, err error) {
	return streams.WriteAllTo(w, c.UserID, c.NodeID, c.ExpiresAt)
}

func (c *NodeContract) ReadFrom(r io.Reader) (n int64, err error) {
	c.UserID = &astral.Identity{}
	c.NodeID = &astral.Identity{}
	return streams.ReadAllFrom(r, c.UserID, c.NodeID, &c.ExpiresAt)
}

func init() {
	astral.DefaultBlueprints.Add(&NodeContract{})
}
