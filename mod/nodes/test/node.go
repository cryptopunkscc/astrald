package nodes

import (
	"io"

	"github.com/cryptopunkscc/astrald/astral"
)

type Node struct {
	identity *astral.Identity
}

var _ astral.Node = &Node{}

func NewNode() (n *Node, err error) {
	i, err := astral.GenerateIdentity()
	if err != nil {
		return
	}
	n = &Node{identity: i}
	return
}

func (t Node) Identity() *astral.Identity { return t.identity }

func (t Node) RouteQuery(ctx *astral.Context, q *astral.Query, w io.WriteCloser) (io.WriteCloser, error) {
	panic("implement me")
}
