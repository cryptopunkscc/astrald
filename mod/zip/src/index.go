package zip

import (
	_zip "archive/zip"
	"errors"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/mod/sets"
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

	var setName = "mod.zip.archive." + zipID.String()
	_, err = mod.sets.CreateSet(setName, sets.TypeSet)
	if err != nil {
		mod.log.Error("error creating set %v: %v", setName, err)
	} else {
		err = mod.sets.AddToUnion(sets.LocalNodeSet, setName)
		if err != nil {
			mod.log.Error("error adding %v to localnode union: %v", setName, err)
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

		mod.content.SetLabel(fileID, filepath.Base(file.Name))

		mod.log.Infov(1, "indexed %s (%v)", file.Name, fileID)

		mod.sets.AddToSet(setName, fileID)
	}

	err = mod.sets.AddToSet(zip.ArchivesSet, zipID)
	if err != nil {
		mod.log.Error("error adding archive to set %v: %v", zip.ArchivesSet, err)
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
