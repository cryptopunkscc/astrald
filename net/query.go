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
}

var _ Query = &basicQuery{}

type basicQuery struct {
	caller id.Identity
	target id.Identity
	query  string
}

func NewQuery(caller id.Identity, target id.Identity, query string) Query {
	return &basicQuery{caller: caller, target: target, query: query}
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
