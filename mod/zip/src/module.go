package zip

import (
	_zip "archive/zip"
	"context"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/content"
	"github.com/cryptopunkscc/astrald/mod/sets"
	"github.com/cryptopunkscc/astrald/mod/shares"
	"github.com/cryptopunkscc/astrald/mod/storage"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/events"
	"github.com/cryptopunkscc/astrald/streams"
	"github.com/cryptopunkscc/astrald/tasks"
	"gorm.io/gorm"
)

const ZipSetType = "zip"

type Module struct {
	config Config
	node   node.Node
	events events.Queue
	log    *log.Logger

	db      *gorm.DB
	content content.Module
	storage storage.Module
	shares  shares.Module
	sets    sets.Module
}

func (mod *Module) Run(ctx context.Context) error {
	return tasks.Group(
		&IndexerService{Module: mod},
	).Run(ctx)
}

func (mod *Module) Open(dataID data.ID, opts *storage.OpenOpts) (storage.Reader, error) {
	if opts == nil {
		opts = &storage.OpenOpts{}
	}

	if !opts.Virtual {
		return nil, storage.ErrNotFound
	}

	if opts.Offset > dataID.Size {
		return nil, storage.ErrInvalidOffset
	}

	var rows []dbContents
	err := mod.db.
		Unscoped().
		Preload("Zip").
		Where("file_id = ?", dataID).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}

	for _, row := range rows {
		r, err := mod.open(row.Zip.DataID, row.Path, opts)
		if err == nil {
			mod.log.Logv(2, "opened %v from %v/%v", dataID, row.Zip.DataID, row.Path)
			return r, nil
		}
	}

	return nil, storage.ErrNotFound
}

func (mod *Module) open(zipID data.ID, path string, opts *storage.OpenOpts) (storage.Reader, error) {
	var zipReaderAt = &readerAt{
		storage: mod.storage,
		dataID:  zipID,
	}

	zipFile, err := _zip.NewReader(zipReaderAt, int64(zipID.Size))
	if err != nil {
		return nil, storage.ErrNotFound
	}

	file, err := zipFile.Open(path)
	if err != nil {
		return nil, err
	}

	if opts.Offset > 0 {
		if err := streams.Skip(file, opts.Offset); err != nil {
			file.Close()
			return nil, err
		}
	}

	return &Reader{File: file, name: "mod.zip"}, err
}
