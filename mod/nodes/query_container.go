package nodes

import (
	"io"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/nodes/frames"
)

type QueryContainer struct {
	TargetID *astral.Identity
	CallerID *astral.Identity
	Query    frames.Query
}

func (c QueryContainer) ObjectType() string { return "nodes.query_container" }

func (c QueryContainer) WriteTo(w io.Writer) (int64, error) {
	return astral.Objectify(&c).WriteTo(w)
}

func (c *QueryContainer) ReadFrom(r io.Reader) (int64, error) {
	return astral.Objectify(c).ReadFrom(r)
}

func NewQueryContainer(q *astral.Query, bufSize int) *QueryContainer {
	return &QueryContainer{
		CallerID: q.Caller,
		TargetID: q.Target,
		Query: frames.Query{
			Nonce:  q.Nonce,
			Buffer: uint32(bufSize),
			Query:  q.Query,
		},
	}
}

func init() {
	_ = astral.Add(&QueryContainer{})
}
