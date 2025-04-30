package fs

import (
	"context"
	"errors"
	"fmt"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/core/assets"
	"github.com/cryptopunkscc/astrald/mod/fs"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/mod/shell"
	"github.com/cryptopunkscc/astrald/object"
	"github.com/cryptopunkscc/astrald/sig"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"os"
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

	repos sig.Map[string, objects.Repository]
	ops   shell.Scope
}

func (mod *Module) Run(ctx *astral.Context) error {
	mod.ctx = ctx

	mod.verifyIndex(ctx)

	<-ctx.Done()
	return nil
}

func (mod *Module) Scope() *shell.Scope {
	return &mod.ops
}

// update updates the index entry for the path. path must be absolute.
func (mod *Module) update(path string) (*object.ID, error) {
	if len(path) == 0 || path[0] != '/' {
		return nil, errors.New("invalid path")
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
		// update if necessary
		err := mod.validate(path)
		if err != nil {
			mod.update(path)
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

// validate checks if the index is up-to-date for the given path
func (mod *Module) validate(path string) error {
	if len(path) == 0 || path[0] != '/' {
		return errors.New("invalid path")
	}

	stat, err := os.Stat(path)
	switch {
	case err != nil:
		return fmt.Errorf("cannot access file: %w", err)
	}

	row, err := mod.db.FindByPath(path)
	if err != nil {
		return errors.New("not indexed")
	}

	if stat.ModTime().After(row.ModTime) {
		return errors.New("file modified")
	}

	return nil
}

func (mod *Module) String() string {
	return fs.ModuleName
}

func resolveFileID(path string) (*object.ID, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	fileID, err := object.Resolve(file)
	if err != nil {
		return nil, err
	}

	return fileID, nil
}
