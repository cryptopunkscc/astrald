package zip

import (
	_zip "archive/zip"
	"context"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/content"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/mod/shares"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/events"
	"github.com/cryptopunkscc/astrald/object"
	"github.com/cryptopunkscc/astrald/tasks"
	"gorm.io/gorm"
)

type Module struct {
	config Config
	node   node.Node
	events events.Queue
	log    *log.Logger

	db      *gorm.DB
	content content.Module
	objects objects.Module
	shares  shares.Module
}

func (mod *Module) Run(ctx context.Context) error {
	return tasks.Group(
		&IndexerService{Module: mod},
	).Run(ctx)
}

func (mod *Module) Open(objectID object.ID, opts *objects.OpenOpts) (objects.Reader, error) {
	if opts == nil {
		opts = &objects.OpenOpts{}
	}

	if !opts.Virtual {
		return nil, objects.ErrNotFound
	}

	if opts.Offset > objectID.Size {
		return nil, objects.ErrInvalidOffset
	}

	var rows []dbContents
	err := mod.db.
		Unscoped().
		Preload("Zip").
		Where("file_id = ?", objectID).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}

	for _, row := range rows {

		r, err := mod.open(row.Zip.DataID, row.Path, row.FileID, opts)
		if err == nil {
			mod.log.Logv(2, "opened %v from %v/%v", objectID, row.Zip.DataID, row.Path)
			return r, nil
		}
	}

	return nil, objects.ErrNotFound
}

func (mod *Module) open(zipID object.ID, path string, fileID object.ID, opts *objects.OpenOpts) (objects.Reader, error) {
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

func (mod *Module) openZip(zipID object.ID) (*_zip.Reader, error) {
	var zipReaderAt = &readerAt{
		objects:  mod.objects,
		objectID: zipID,
	}

	zipFile, err := _zip.NewReader(zipReaderAt, int64(zipID.Size))
	return zipFile, err
}
