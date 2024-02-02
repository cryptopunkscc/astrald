package zip

import (
	_zip "archive/zip"
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
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

	allArchived sets.Union
}

func (mod *Module) Run(ctx context.Context) error {
	return tasks.Group(
		&IndexerService{Module: mod},
	).Run(ctx)
}

func (mod *Module) Read(dataID data.ID, opts *storage.ReadOpts) (storage.DataReader, error) {
	if opts == nil {
		opts = &storage.ReadOpts{}
	}

	if !opts.Virtual {
		return nil, storage.ErrNotFound
	}

	if opts.Offset > dataID.Size {
		return nil, storage.ErrInvalidOffset
	}

	var zipRows, err = mod.dbFindByFileID(dataID)
	if err != nil {
		return nil, err
	}
	if len(zipRows) == 0 {
		return nil, storage.ErrNotFound
	}

	var zipRow = zipRows[0]

	zipID := zipRow.Zip.DataID
	if err != nil {
		return nil, err
	}

	var zipReaderAt = &readerAt{
		storage: mod.storage,
		dataID:  zipID,
	}

	zipFile, err := _zip.NewReader(zipReaderAt, int64(zipID.Size))
	if err != nil {
		return nil, storage.ErrNotFound
	}

	file, err := zipFile.Open(zipRow.Path)
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

// Authorize authorizes access if the dataID is contained within a zip file that the identity has access to.
func (mod *Module) Authorize(identity id.Identity, dataID data.ID) error {
	var rows []*dbContents

	var tx = mod.db.
		Unscoped().
		Preload("Zip").
		Where("file_id = ?", dataID).
		Find(&rows)
	if tx.Error != nil {
		return shares.ErrDenied
	}

	for _, row := range rows {
		if row.Zip == nil {
			mod.log.Errorv(1, "db row for file %v has null reference to zip", dataID)
			continue
		}

		zipID := row.Zip.DataID

		// sanity check to avoid infitite loops
		if zipID.IsEqual(dataID) {
			continue
		}

		if mod.shares.Authorize(identity, zipID) == nil {
			return nil
		}
	}

	return shares.ErrDenied
}
