package zip

import (
	_zip "archive/zip"
	"errors"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/mod/index"
	"github.com/cryptopunkscc/astrald/mod/zip"
	"path/filepath"
)

func (mod *Module) Index(zipID data.ID, reindex bool) error {
	if mod.isIndexed(zipID) && !reindex {
		return errors.New("already indexed")
	}

	mod.log.Logv(1, "indexing %v", zipID)

	r := &readerAt{
		storage: mod.storage,
		dataID:  zipID,
	}

	reader, err := _zip.NewReader(r, int64(zipID.Size))
	if err != nil {
		return err
	}

	var indexName = "mod.zip.archive." + zipID.String()
	_, err = mod.index.CreateIndex(indexName, index.TypeSet)
	if err != nil {
		mod.log.Error("error creating index %v: %v", indexName, err)
	} else {
		err = mod.index.AddToUnion(index.LocalNodeUnionName, indexName)
		if err != nil {
			mod.log.Error("error adding %v to localnode union: %v", indexName, err)
		}
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

		mod.db.Create(&dbZipContent{
			ZipID:  zipID.String(),
			Path:   file.Name,
			FileID: fileID.String(),
		})

		mod.data.SetLabel(fileID, filepath.Base(file.Name))

		mod.log.Infov(1, "indexed %s (%v)", file.Name, fileID)

		mod.index.AddToSet(indexName, fileID)
	}

	err = mod.index.AddToSet(zip.ArchivesIndexName, zipID)
	if err != nil {
		mod.log.Error("error adding archive to %v index: %v", zip.ArchivesIndexName, err)
	}

	mod.events.Emit(zip.EventArchiveIndexed{DataID: zipID})

	return nil
}

func (mod *Module) isIndexed(dataID data.ID) bool {
	var count int64
	tx := mod.db.Model(&dbZipContent{}).Where("zip_id = ?", dataID.String()).Count(&count)
	if tx.Error != nil {
		mod.log.Errorv(2, "database error: %v", tx.Error)
		return false
	}

	return count > 0
}
