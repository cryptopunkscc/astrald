package archives

import (
	_zip "archive/zip"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/object"
	"io"
)

func (mod *Module) OpenObject(ctx *astral.Context, objectID *object.ID) (io.ReadCloser, error) {
	if !ctx.Zone().Is(astral.ZoneVirtual) {
		return nil, astral.ErrZoneExcluded
	}

	if objects.IsOffsetLimitValid(objectID, 0, 0) {
		return nil, objects.ErrOutOfBounds
	}

	var rows []dbEntry
	err := mod.db.
		Unscoped().
		Preload("Parent").
		Where("object_id = ?", objectID).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}

	for _, row := range rows {
		r, err := mod.open(row.Parent.ObjectID, row.Path, row.ObjectID)
		if err == nil {
			mod.log.Logv(2, "opened %v from %v/%v", objectID, row.Parent.ObjectID, row.Path)
			return r, nil
		}
	}

	return nil, objects.ErrNotFound
}

func (mod *Module) open(zipID *object.ID, path string, fileID *object.ID) (io.ReadCloser, error) {
	zipFile, err := mod.openZip(zipID)
	if err != nil {
		return nil, objects.ErrNotFound
	}

	var r = &contentReader{
		zip:      zipFile,
		path:     path,
		objectID: fileID,
	}

	err = r.open()

	return r, err
}

func (mod *Module) openZip(objectID *object.ID) (*_zip.Reader, error) {
	var r = &readerAt{
		identity: mod.node.Identity(),
		objects:  mod.Objects,
		objectID: objectID,
	}

	zipFile, err := _zip.NewReader(r, int64(objectID.Size))
	return zipFile, err
}
