package status

import (
	"errors"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/dir"
	"strings"
)

var _ dir.Resolver = &Module{}

func (mod *Module) ResolveIdentity(s string) (*astral.Identity, error) {
	s, found := strings.CutPrefix(s, aliasPrefix)
	if !found {
		return nil, errors.New("not found")
	}

	for _, v := range mod.Cache().Clone() {
		if string(v.Status.Alias) == s {
			return v.Identity, nil
		}
	}

	return nil, errors.New("not found")
}

func (mod *Module) DisplayName(identity *astral.Identity) string {
	return ""
}
