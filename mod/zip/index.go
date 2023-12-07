package zip

import (
	"archive/zip"
	"errors"
	"github.com/cryptopunkscc/astrald/data"
	storage "github.com/cryptopunkscc/astrald/mod/storage/api"
	"path/filepath"
	"time"
)

func (mod *Module) indexZip(zipID data.ID, reindex bool) error {
	if mod.isIndexed(zipID) && !reindex {
		return errors.New("already indexed")
	}

	mod.log.Logv(1, "indexing %v", zipID)

	r := &readerAt{
		storage: mod.storage,
		dataID:  zipID,
	}

	reader, err := zip.NewReader(r, int64(zipID.Size))
	if err != nil {
		return err
	}

	for _, file := range reader.File {
		f, err := file.Open()
		if err != nil {
			mod.log.Errorv(1, "open %v: %v", file.Name, err)
			continue
		}
		defer f.Close()

		fileID, err := data.ResolveAll(f)
		if err != nil {
			mod.log.Errorv(1, "resolve %v: %v", file.Name, err)
			continue
		}

		indexedAt := time.Now()

		mod.db.Create(&dbZipContent{
			ZipID:     zipID.String(),
			Path:      file.Name,
			FileID:    fileID.String(),
			IndexedAt: indexedAt,
		})

		mod.data.SetLabel(fileID, filepath.Base(file.Name))

		mod.log.Infov(1, "indexed %s (%v)", file.Name, fileID)

		mod.events.Emit(storage.EventDataAdded{
			ID:        fileID,
			IndexedAt: indexedAt,
		})

	}

	return nil
}

func (mod *Module) isIndexed(dataID data.ID) bool {
	var count int64
	tx := mod.db.Model(&dbZipContent{}).Where("zip_id = ?", dataID.String()).Count(&count)
	if tx.Error != nil {
		return false
	}

	return count > 0
}