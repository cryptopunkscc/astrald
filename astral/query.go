package astral

import (
	"github.com/cryptopunkscc/astrald/sig"
)

const OriginNetwork = "network"
const OriginLocal = "local"

type Query struct {
	Nonce  Nonce
	Caller *Identity
	Target *Identity
	Query  string
	Extra  sig.Map[string, any]
}

// NewQuery returns a Query instance with a random Nonce.
func NewQuery(caller *Identity, target *Identity, query string) *Query {
	return &Query{
		Nonce:  NewNonce(),
		Caller: caller,
		Target: target,
		Query:  query,
	}
}

func (q *Query) IsNetwork() bool {
	o, ok := q.Extra.Get("origin")
	return ok && (o == OriginNetwork)
}

func (q *Query) IsLocal() bool {
	o, _ := q.Extra.Get("origin")
	return o == "" || o == OriginLocal
}
