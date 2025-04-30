package media

import (
	"bytes"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/mod/objects/mem"
	"github.com/cryptopunkscc/astrald/object"
	"github.com/dhowden/tag"
)

type Repository struct {
	mod *Module
	objects.NilRepository
}

var _ objects.Repository = &Repository{}

func (repo *Repository) Label() string {
	return "Media covers"
}

func (repo *Repository) Read(ctx *astral.Context, objectID *object.ID, offset int64, limit int64) (objects.Reader, error) {
	if !ctx.Zone().Is(astral.ZoneVirtual) {
		return nil, astral.ErrZoneExcluded
	}

	containerID, err := repo.mod.db.FindAudioContainerID(objectID)
	if err != nil {
		return nil, err
	}
	if containerID.IsZero() {
		return nil, objects.ErrNotFound
	}

	r, err := repo.mod.Objects.Root().Read(ctx, containerID, 0, 0)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	audioTag, err := tag.ReadFrom(r)
	if err != nil {
		return nil, err
	}

	if audioTag.Picture() == nil {
		return nil, objects.ErrNotFound
	}

	pic := audioTag.Picture().Data

	actualID, _ := object.Resolve(bytes.NewReader(pic))
	if !actualID.IsEqual(objectID) {
		return nil, objects.ErrNotFound
	}

	return mem.NewReader(pic[:]), nil
}
