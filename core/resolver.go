package core

import (
	"fmt"
	"github.com/cryptopunkscc/astrald/id"
	"github.com/cryptopunkscc/astrald/node"
)

const ZeroIdentity = "<anyone>"

var _ node.ResolverEngine = &Resolver{}

type Resolver struct {
	node      node.Node
	resolvers []node.Resolver
}

func NewResolver(n node.Node) *Resolver {
	return &Resolver{
		node:      n,
		resolvers: make([]node.Resolver, 0),
	}
}

func (c *Resolver) Resolve(s string) (id.Identity, error) {
	if s == "" || s == "anyone" {
		return id.Identity{}, nil
	}

	if s == "localnode" {
		return c.node.Identity(), nil
	}

	if identity, err := id.ParsePublicKeyHex(s); err == nil {
		return identity, nil
	}

	for _, r := range c.resolvers {
		if i, err := r.Resolve(s); err == nil {
			return i, nil
		}
	}

	return id.Identity{}, fmt.Errorf("unknown identity: %s", s)
}

func (c *Resolver) DisplayName(identity id.Identity) string {
	if identity.IsZero() {
		return ZeroIdentity
	}

	for _, r := range c.resolvers {
		if s := r.DisplayName(identity); len(s) > 0 {
			return s
		}
	}

	return identity.Fingerprint()
}

func (c *Resolver) AddResolver(r node.Resolver) error {
	c.resolvers = append(c.resolvers, r)
	return nil
}
