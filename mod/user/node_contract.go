package user

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/streams"
	"io"
	"time"
)

type NodeContract struct {
	UserID    *astral.Identity
	NodeID    *astral.Identity
	ExpiresAt astral.Time
}

var _ astral.Object = &NodeContract{}

func (NodeContract) ObjectType() string {
	return "mod.users.node_contract"
}

func (c *NodeContract) IsExpired() bool {
	return time.Now().After(c.ExpiresAt.Time())
}

func (c NodeContract) WriteTo(w io.Writer) (n int64, err error) {
	return streams.WriteAllTo(w, c.UserID, c.NodeID, astral.Time(c.ExpiresAt))
}

func (c *NodeContract) ReadFrom(r io.Reader) (n int64, err error) {
	c.UserID = &astral.Identity{}
	c.NodeID = &astral.Identity{}
	return streams.ReadAllFrom(r, c.UserID, c.NodeID, &c.ExpiresAt)
}
