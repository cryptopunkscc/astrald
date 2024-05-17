package objects

import (
	"context"
	"github.com/cryptopunkscc/astrald/lib/desc"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/object"
)

func (mod *Module) Describe(ctx context.Context, objectID object.ID, opts *desc.Opts) []*desc.Desc {
	var describers []desc.Describer[object.ID]

	for _, d := range mod.describers.Clone() {
		describers = append(describers, d)
	}

	return desc.Collect(ctx, objectID, opts, describers...)
}

func (mod *Module) AddDescriber(describer objects.Describer) error {
	return mod.describers.Add(describer)
}
