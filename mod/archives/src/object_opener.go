package archives

import (
	_zip "archive/zip"
	"context"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/object"
)

func (mod *Module) OpenObject(ctx *context.Context, objectID object.ID, opts *objects.OpenOpts) (objects.Reader, error) {
	if opts == nil {
		opts = &objects.OpenOpts{}
	}

	if !opts.Zone.Is(astral.ZoneVirtual) {
		return nil, astral.ErrZoneExcluded
	}

	if opts.Offset > objectID.Size {
		return nil, objects.ErrInvalidOffset
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
		r, err := mod.open(row.Parent.ObjectID, row.Path, row.ObjectID, opts)
		if err == nil {
			mod.log.Logv(2, "opened %v from %v/%v", objectID, row.Parent.ObjectID, row.Path)
			return r, nil
		}
	}

	return nil, objects.ErrNotFound
}

func (mod *Module) open(zipID object.ID, path string, fileID object.ID, opts *objects.OpenOpts) (objects.Reader, error) {
	zipFile, err := mod.openZip(zipID, opts)
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

func (mod *Module) openZip(objectID object.ID, opts *objects.OpenOpts) (*_zip.Reader, error) {
	var r = &readerAt{
		objects:  mod.Objects,
		objectID: objectID,
		opts:     opts,
	}

	zipFile, err := _zip.NewReader(r, int64(objectID.Size))
	return zipFile, err
}
