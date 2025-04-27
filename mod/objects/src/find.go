package objects

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/object"
)

func (mod *Module) AddFinder(finder objects.Finder) error {
	return mod.finders.Add(finder)
}

func (mod *Module) Find(ctx *astral.Context, objectID *object.ID) (providers []*astral.Identity) {
	var unique = map[string]struct{}{}

	for _, finder := range mod.finders.Clone() {
		for _, i := range finder.FindObject(ctx, objectID, astral.DefaultScope()) {
			if _, found := unique[i.String()]; found {
				continue
			}
			unique[i.String()] = struct{}{}
			providers = append(providers, i)
		}
	}

	return
}
