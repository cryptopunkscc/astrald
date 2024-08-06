package archives

import (
	"context"
	"fmt"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/archives"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/object"
)

type entryFunc func(*archives.Entry)

func (mod *Module) Index(ctx context.Context, objectID object.ID, opts *objects.OpenOpts) (archive *archives.Archive, err error) {
	mod.mu.Lock()
	defer mod.mu.Unlock()

	if cached := mod.getCache(objectID); cached != nil {
		return cached, nil
	}

	mod.log.Logv(1, "indexing zip %v", objectID)
	archive, err = mod.scan(ctx, objectID, opts, func(entry *archives.Entry) {
		mod.log.Infov(1, "scanned %v (%s)", entry.ObjectID, entry.Path)
	})
	if err != nil {
		return
	}

	err = mod.setCache(objectID, archive)

	mod.events.Emit(archives.EventArchiveIndexed{ObjectID: objectID, Archive: archive})
	for _, entry := range archive.Entries {
		mod.events.Emit(objects.EventDiscovered{
			ObjectID: entry.ObjectID,
			Zone:     astral.ZoneVirtual,
		})
	}

	return
}

func (mod *Module) scan(ctx context.Context, objectID object.ID, opts *objects.OpenOpts, postScan entryFunc) (archive *archives.Archive, err error) {
	reader, err := mod.openZip(objectID, opts)
	if err != nil {
		return nil, fmt.Errorf("error reading zip file: %w", err)
	}

	archive = &archives.Archive{
		Comment: reader.Comment,
		Format:  "zip",
	}

	for _, file := range reader.File {
		if file.FileInfo().IsDir() {
			continue
		}

		f, err := file.Open()
		if err != nil {
			mod.log.Errorv(1, "open %v: %v", file.Name, err)
			continue
		}

		fileID, err := object.Resolve(f)
		f.Close()
		if err != nil {
			mod.log.Errorv(1, "resolve %v: %v", file.Name, err)
			continue
		}

		entry := &archives.Entry{
			ObjectID: fileID,
			Path:     file.Name,
			Comment:  file.Comment,
			Modified: file.Modified,
		}

		archive.Entries = append(archive.Entries, entry)

		if postScan != nil {
			postScan(entry)
		}

		select {
		case <-ctx.Done():
			return archive, ctx.Err()
		default:
		}
	}

	return
}

func (mod *Module) Forget(objectID object.ID) error {
	return mod.clearCache(objectID)
}

func (mod *Module) getCache(objectID object.ID) (archive *archives.Archive) {
	var row dbArchive

	err := mod.db.
		Where("object_id = ?", objectID).
		Preload("Entries").
		First(&row).
		Error
	if err != nil {
		return
	}

	archive = &archives.Archive{
		Comment: row.Comment,
		Format:  row.Format,
	}

	for _, e := range row.Entries {
		archive.Entries = append(archive.Entries, &archives.Entry{
			ObjectID: e.ObjectID,
			Path:     e.Path,
			Comment:  e.Comment,
			Modified: e.Modified,
		})
	}

	return
}

func (mod *Module) clearCache(objectID object.ID) (err error) {
	var id int
	err = mod.db.
		Model(&dbArchive{}).
		Select("id").
		Where("object_id = ?", objectID).
		First(&id).
		Error
	if err != nil {
		return
	}
	err = mod.db.
		Where("parent_id = ?", id).
		Delete(&dbEntry{}).
		Error
	if err != nil {
		return
	}
	return mod.db.
		Where("object_id = ?", objectID).
		Delete(&dbArchive{}).
		Error
}

func (mod *Module) setCache(objectID object.ID, archive *archives.Archive) error {
	mod.clearCache(objectID)

	row := dbArchive{
		ObjectID: objectID,
		Comment:  archive.Comment,
		Format:   archive.Format,
	}

	for _, entry := range archive.Entries {
		row.Entries = append(row.Entries, dbEntry{
			ObjectID: entry.ObjectID,
			Path:     entry.Path,
			Comment:  entry.Comment,
			Modified: entry.Modified,
		})
	}

	return mod.db.Create(&row).Error
}
