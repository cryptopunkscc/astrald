package core

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/sig"
)

type QueryModifier struct {
	query   *astral.Query
	blocked sig.Value[error]
}

func (q *QueryModifier) Query() *astral.Query {
	return q.query
}

func (q *QueryModifier) Block(err error) {
	q.blocked.Swap(nil, err)
}

func (q *QueryModifier) Attach(object astral.Object) {
	q.query.Extra.Set(nodes.ExtraCallerProof, object)
}

func (q *QueryModifier) AddRelay(identity *astral.Identity) {
	if l, ok := q.query.Extra.Get(nodes.ExtraRelayVia); ok {
		if l, ok := l.([]*astral.Identity); ok {
			q.query.Extra.Replace(nodes.ExtraRelayVia, append(l, identity))
			return
		}
	}

	q.query.Extra.Set(nodes.ExtraRelayVia, []*astral.Identity{identity})
}

func (q *QueryModifier) SetCaller(identity *astral.Identity) {
	q.query.Caller = identity
}

func (q *QueryModifier) SetTarget(identity *astral.Identity) {
	q.query.Target = identity
}
