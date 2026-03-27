package nearby

import (
	"errors"
	"strings"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/dir"
)

var _ dir.Resolver = &Module{}

func (mod *Module) ResolveIdentity(s string) (*astral.Identity, error) {
	s, found := strings.CutPrefix(s, aliasPrefix)
	if !found {
		return nil, errors.New("not found")
	}

	for _, v := range mod.Cache().Clone() {
		if v.Identity == nil {
			continue
		}

		alias, ok := astral.First[*dir.Alias](v.Status.Attachments.Objects())
		if ok && alias != nil && alias.String() == s {
			return v.Identity, nil
		}

	}

	return nil, errors.New("not found")
}

func (mod *Module) DisplayName(identity *astral.Identity) string {
	return ""
}
