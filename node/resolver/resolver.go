package resolver

import (
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
)

type Resolver interface {
	Resolve(s string) (id.Identity, error)
	DisplayName(identity id.Identity) string
}

var _ Resolver = &CoreResolver{}

type Node interface {
	Identity() id.Identity
	Alias() string
}

type CoreResolver struct {
	node      Node
	resolvers []Resolver
}

func NewCoreResolver(node Node) *CoreResolver {
	return &CoreResolver{
		node:      node,
		resolvers: make([]Resolver, 0),
	}
}

func (c *CoreResolver) Resolve(s string) (id.Identity, error) {
	if s == "localnode" || s == c.node.Alias() {
		return c.node.Identity(), nil
	}

	for _, r := range c.resolvers {
		if i, err := r.Resolve(s); err == nil {
			return i, nil
		}
	}

	return id.Identity{}, errors.New("unknown identity")
}

func (c *CoreResolver) DisplayName(identity id.Identity) string {
	if identity.IsEqual(c.node.Identity()) {
		return c.node.Alias()
	}

	for _, r := range c.resolvers {
		if s := r.DisplayName(identity); len(s) > 0 {
			return s
		}
	}

	return identity.Fingerprint()
}

func (c *CoreResolver) AddResolver(r Resolver) error {
	c.resolvers = append(c.resolvers, r)
	return nil
}
