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

func GetExtra[T any](q *Query, key string) (v T, ok bool) {
	if a, ok := q.Extra.Get(key); ok {
		v, ok = a.(T)
	}
	return
}
