package media

import (
	"bytes"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/mod/objects/mem"
	"github.com/cryptopunkscc/astrald/object"
	"github.com/dhowden/tag"
)

func (mod *Module) OpenObject(ctx *astral.Context, objectID *object.ID) (objects.Reader, error) {
	if !ctx.Zone().Is(astral.ZoneVirtual) {
		return nil, astral.ErrZoneExcluded
	}
	var parentID = mod.getParentID(objectID)
	if parentID.IsZero() {
		return nil, objects.ErrNotFound
	}

	r, err := mod.Objects.Root().Read(ctx, parentID, 0, 0)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	meta, err := tag.ReadFrom(r)
	if err != nil {
		return nil, err
	}

	if meta.Picture() == nil {
		return nil, objects.ErrNotFound
	}

	pic := meta.Picture().Data

	actualID, _ := object.Resolve(bytes.NewReader(pic))
	if !actualID.IsEqual(objectID) {
		return nil, objects.ErrNotFound
	}

	return mem.NewReader(pic[:]), nil
}
