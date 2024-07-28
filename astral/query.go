package astral

import (
	"github.com/cryptopunkscc/astrald/id"
)

const OriginNetwork = "network"
const OriginLocal = "local"

type Query struct {
	Nonce  Nonce
	Caller id.Identity
	Target id.Identity
	Query  string
}

// NewQuery returns a Query instance with a random Nonce.
func NewQuery(caller id.Identity, target id.Identity, query string) *Query {
	return &Query{
		Nonce:  NewNonce(),
		Caller: caller,
		Target: target,
		Query:  query,
	}
}
