package net

import (
	"github.com/cryptopunkscc/astrald/auth/id"
)

const OriginNetwork = "network"
const OriginLocal = "local"

type Query interface {
	Caller() id.Identity
	Target() id.Identity
	Query() string
	Origin() string
}

var _ Query = &query{}

type query struct {
	caller id.Identity
	target id.Identity
	query  string
	origin string
}

func NewQuery(caller id.Identity, target id.Identity, q string) Query {
	return &query{caller: caller, target: target, query: q, origin: OriginLocal}
}

func NewQueryOrigin(caller id.Identity, target id.Identity, q string, origin string) Query {
	return &query{caller: caller, target: target, query: q, origin: origin}
}

func (q *query) Caller() id.Identity {
	return q.caller
}

func (q *query) Target() id.Identity {
	return q.target
}

func (q *query) Query() string {
	return q.query
}

func (q *query) Origin() string {
	return q.origin
}
