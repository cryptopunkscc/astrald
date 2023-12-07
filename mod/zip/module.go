package zip

import (
	"archive/zip"
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
	_data "github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/log"
	admin "github.com/cryptopunkscc/astrald/mod/admin/api"
	data "github.com/cryptopunkscc/astrald/mod/data/api"
	storage "github.com/cryptopunkscc/astrald/mod/storage/api"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/events"
	"github.com/cryptopunkscc/astrald/streams"
	"github.com/cryptopunkscc/astrald/tasks"
	"gorm.io/gorm"
	"io"
	"time"
)

type Module struct {
	config Config
	node   node.Node
	events events.Queue
	log    *log.Logger

	db      *gorm.DB
	data    data.API
	storage storage.API
}

func (mod *Module) Run(ctx context.Context) error {
	// inject admin command
	if adm, _ := mod.node.Modules().Find("admin").(admin.API); adm != nil {
		adm.AddCommand("zip", NewAdmin(mod))
	}

	mod.data, _ = mod.node.Modules().Find("data").(data.API)
	if mod.data == nil {
		return errors.New("data module not found")
	}

	mod.storage, _ = mod.node.Modules().Find("storage").(storage.API)
	if mod.storage == nil {
		return errors.New("storage module not found")
	}

	mod.storage.Data().AddIndexer(mod)
	defer mod.storage.Data().RemoveIndexer(mod)
	mod.storage.Data().AddReader(mod)
	defer mod.storage.Data().RemoveReader(mod)
	mod.storage.Access().AddAccessVerifier(mod)
	defer mod.storage.Access().RemoveAccessVerifier(mod)

	return tasks.Group(
		&Service{Module: mod},
	).Run(ctx)
}

func (mod *Module) Read(dataID _data.ID, opts *storage.ReadOpts) (io.ReadCloser, error) {
	if opts == nil {
		opts = &storage.ReadOpts{}
	}

	if opts.Offset > dataID.Size {
		return nil, storage.ErrInvalidOffset
	}

	var zipRow = mod.dbFindByID(dataID)
	if zipRow == nil {
		return nil, storage.ErrNotFound
	}

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

	zipFile, err := zip.NewReader(zipReaderAt, int64(zipID.Size))
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

	return file, err
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

func (mod *Module) IndexSince(time time.Time) []storage.DataInfo {
	var list []storage.DataInfo
	var rows []*dbZipContent

	mod.db.Where("indexed_at > ?", time).Order("indexed_at").Find(&rows)

	for _, row := range rows {
		dataID, err := _data.Parse(row.FileID)
		if err != nil {
			continue
		}

		list = append(list, storage.DataInfo{
			ID:        dataID,
			IndexedAt: row.IndexedAt,
		})
	}

	return list
}
