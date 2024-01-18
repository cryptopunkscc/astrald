package zip

import (
	_zip "archive/zip"
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	_data "github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/data"
	"github.com/cryptopunkscc/astrald/mod/index"
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
	data    data.Module
	storage storage.Module
	index   index.Module
}

func (mod *Module) Run(ctx context.Context) error {
	return tasks.Group(
		&IndexerService{Module: mod},
	).Run(ctx)
}

func (mod *Module) Read(dataID _data.ID, opts *storage.ReadOpts) (storage.DataReader, error) {
	if opts == nil {
		opts = &storage.ReadOpts{}
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

	if opts.NoVirtual {
		return nil, storage.ErrNoVirtual
	}

	zipID, err := _data.Parse(zipRow.ZipID)
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

func (mod *Module) Verify(identity id.Identity, dataID _data.ID) bool {
	var rows []*dbZipContent

	tx := mod.db.Where("file_id = ?", dataID.String()).Find(&rows)
	if tx.Error != nil {
		return false
	}

	for _, row := range rows {
		zipID, err := _data.Parse(row.ZipID)
		if err != nil {
			continue
		}

		if mod.storage.Access().Verify(identity, zipID) {
			return true
		}
	}

	return false
}
