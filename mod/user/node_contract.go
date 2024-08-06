package user

import (
	"crypto/sha256"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/streams"
	"io"
	"time"
)

var _ astral.Object = &NodeContract{}

type NodeContract struct {
	UserID    *astral.Identity
	NodeID    *astral.Identity
	ExpiresAt astral.Time
}

func (NodeContract) ObjectType() string {
	return "mod.users.node_contract"
}

func (c *NodeContract) Hash() []byte {
	var hash = sha256.New()
	_, err := c.WriteTo(hash)
	if err != nil {
		return nil
	}
	return hash.Sum(nil)
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
