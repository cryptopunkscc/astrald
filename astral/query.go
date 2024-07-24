package astral

import (
	"github.com/cryptopunkscc/astrald/id"
)

const OriginNetwork = "network"
const OriginLocal = "local"

type Query interface {
	Nonce() Nonce
	Caller() id.Identity
	Target() id.Identity
	Query() string
}

var _ Query = &basicQuery{}

type basicQuery struct {
	nonce  Nonce
	caller id.Identity
	target id.Identity
	query  string
}

func NewQuery(caller id.Identity, target id.Identity, query string) Query {
	return NewQueryNonce(caller, target, query, NewNonce())
}

func NewQueryNonce(caller id.Identity, target id.Identity, query string, nonce Nonce) Query {
	return &basicQuery{
		nonce:  nonce,
		caller: caller,
		target: target,
		query:  query,
	}
}

func (q *basicQuery) Nonce() Nonce {
	return q.nonce
}

func (q *basicQuery) Caller() id.Identity {
	return q.caller
}

func (q *basicQuery) Target() id.Identity {
	return q.target
}

func (q *basicQuery) Query() string {
	return q.query
}
