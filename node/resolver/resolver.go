package resolver

import (
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/node/tracker"
)

const ZeroIdentity = "<anyone>"

type Resolver interface {
	Resolve(s string) (id.Identity, error)
	DisplayName(identity id.Identity) string
}

var _ Resolver = &CoreResolver{}

type Node interface {
	Identity() id.Identity
	Tracker() tracker.Tracker
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
	if s == "" || s == "anyone" {
		return id.Identity{}, nil
	}

	if s == "localnode" {
		return c.node.Identity(), nil
	}

	if identity, err := id.ParsePublicKeyHex(s); err == nil {
		return identity, nil
	}

	if identity, err := c.node.Tracker().IdentityByAlias(s); err == nil {
		return identity, nil
	}

	for _, r := range c.resolvers {
		if i, err := r.Resolve(s); err == nil {
			return i, nil
		}
	}

	return id.Identity{}, errors.New("unknown identity")
}

func (c *CoreResolver) DisplayName(identity id.Identity) string {
	if identity.IsZero() {
		return ZeroIdentity
	}

	if alias, err := c.node.Tracker().GetAlias(identity); err == nil {
		return alias
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
