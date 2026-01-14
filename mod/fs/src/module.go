package fs

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/core/assets"
	"github.com/cryptopunkscc/astrald/mod/fs"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/mod/shell"
	"github.com/cryptopunkscc/astrald/sig"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var _ fs.Module = &Module{}

const workers = 1
const updatesLen = 4096

type Module struct {
	Deps
	config Config
	node   astral.Node
	assets assets.Assets
	log    *log.Logger
	db     *DB
	ctx    context.Context

	fileIndexer *FileIndexer

	repos sig.Map[string, objects.Repository]
	ops   shell.Scope
}

func (mod *Module) Run(ctx *astral.Context) error {
	mod.ctx = ctx

	mod.verifyIndex(ctx)

	<-ctx.Done()

	_ = mod.fileIndexer.Close()
	return nil
}

func (mod *Module) Scope() *shell.Scope {
	return &mod.ops
}

// updateDbIndex updates the database index for the given absolute filesystem path.
// operation is considered heavy due to resolution file bytes into an object ID
// on duplicate path entries the existing entry is updated
func (mod *Module) updateDbIndex(path string) (*astral.ObjectID, error) {
	if len(path) == 0 || path[0] != '/' {
		return nil, fs.ErrInvalidPath
	}

	stat, err := os.Stat(path)
	if err != nil {
		err = mod.db.DeleteByPath(path)
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
		case err != nil:
			mod.log.Errorv(2, "db error: %v", err)
		}
		return nil, nil
	}

	objectID, err := resolveFileID(path)
	if err != nil {
		return nil, err
	}

	updated := &dbLocalFile{
		Path:    path,
		DataID:  objectID,
		ModTime: stat.ModTime(),
	}

	err = mod.db.
		Clauses(
			clause.OnConflict{
				DoUpdates: clause.AssignmentColumns([]string{"data_id", "mod_time"}),
			}).
		Save(updated).
		Error

	return updated.DataID, err
}

func (mod *Module) verifyIndex(ctx *astral.Context) {
	mod.log.Log("verifying index...")

	var updated, total int
	err := mod.db.EachPath(func(path string) error {
		total++
		// updateDbIndex if necessary
		err := mod.checkIndexEntry(path)
		if err != nil {
			mod.updateDbIndex(path)
			updated++
		}

		// check context
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		return nil
	})

	// log
	if err != nil {
		mod.log.Error("index verification finished with error (updated %v/%v): %v", updated, total, err)
	} else {
		mod.log.Info("index verification finished (updated %v/%v)", updated, total)
	}
}

// checkIndexEntry checks if the index is up-to-date for the given path
func (mod *Module) checkIndexEntry(path string) error {
	if len(path) == 0 || path[0] != '/' {
		return fs.ErrInvalidPath
	}

	stat, err := os.Stat(path)
	switch {
	case err != nil:
		return fmt.Errorf("%w: %w", fs.ErrNotFound, err)
	}

	row, err := mod.db.FindByPath(path)
	if err != nil {
		return fs.ErrNotIndexed
	}

	if stat.ModTime().After(row.ModTime) {
		return fs.ErrFileModified
	}

	return nil
}

func (mod *Module) String() string {
	return fs.ModuleName
}

func resolveFileID(path string) (*astral.ObjectID, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	fileID, err := astral.Resolve(file)
	if err != nil {
		return nil, err
	}

	return fileID, nil
}
