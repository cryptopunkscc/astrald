package main

import (
	"errors"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/dir"
)

// resolver implements dir.Resolver from cache
type resolver struct {
	aliasMap *dir.AliasMap
	revMap   map[string]string
}

var _ dir.Resolver = &resolver{}

func newResolver(aliasMap *dir.AliasMap) *resolver {
	r := &resolver{
		aliasMap: aliasMap,
		revMap:   make(map[string]string),
	}

	for k, v := range aliasMap.Aliases {
		r.revMap[v.String()] = k
	}

	return r
}

func (r resolver) ResolveIdentity(s string) (*astral.Identity, error) {
	id, ok := r.aliasMap.Aliases[s]
	if !ok {
		return nil, errors.New("resolution failed")
	}
	return id, nil
}

func (r resolver) DisplayName(identity *astral.Identity) string {
	name := r.revMap[identity.String()]
	if name == "" {
		return identity.String()
	}
	return name
}
