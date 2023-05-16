package contacts

import (
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/node/resolver"
)

var _ resolver.Resolver = &Resolver{}

type Resolver struct {
	mod *Module
}

func (r *Resolver) Resolve(s string) (id.Identity, error) {
	if identity, err := id.ParsePublicKeyHex(s); err == nil {
		return identity, nil
	}

	if node, err := r.mod.FindByAlias(s); err == nil {
		return node.Identity, nil
	}
	return id.Identity{}, errors.New("not found")
}

func (r *Resolver) DisplayName(identity id.Identity) string {
	if node, err := r.mod.Find(identity); err == nil {
		return node.Alias
	}

	return ""
}
