package fs

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync/atomic"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/sig"
	"gorm.io/gorm"
)

type IndexEvent struct {
	Path     string
	ObjectID *astral.ObjectID
}

type Indexer struct {
	mod *Module
	log *log.Logger

	workqueue chan string
	pending   sig.Set[string]

	roots sig.Set[string]

	events          *sig.Queue[IndexEvent]
	subscriberCount atomic.Int32
}

func NewIndexer(mod *Module, workers int) *Indexer {
	indexer := &Indexer{
		mod:       mod,
		log:       mod.log,
		workqueue: make(chan string, 1024),
		events:    &sig.Queue[IndexEvent]{},
	}

	for i := 0; i < workers; i++ {
		go indexer.worker(mod.ctx)
	}

	return indexer
}

func (indexer *Indexer) worker(ctx *astral.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case path := <-indexer.workqueue:
			indexer.pending.Remove(path)

			var found bool
			for _, root := range indexer.roots.Clone() {
				if pathUnderRoot(path, root) {
					found = true
					break
				}
			}

			if !found {
				err := indexer.mod.db.DeleteByPath(path)
				if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
					indexer.log.Error("indexer: delete path %v: %v", path, err)
				}

				// We are no longer interested in this path, no point checking it
				continue
			}

			if err := indexer.checkAndFix(path); err != nil {
				indexer.log.Error("indexer: checkAndFix %v: %v", path, err)
			}
		}
	}

}

func (indexer *Indexer) invalidate(path string) (err error) {
	err = indexer.mod.db.InvalidatePath(path)
	if err != nil {
		return fmt.Errorf("db invalidate: %w", err)
	}

	err = indexer.pending.Add(path)
	if err == nil {
		indexer.workqueue <- path
	}

	return
}

func (indexer *Indexer) checkAndFix(path string) error {
	if len(path) == 0 || path[0] != '/' {
		return errors.New("invalid path")
	}

	stat, err := os.Stat(path)
	if err != nil {
		err = indexer.mod.db.DeleteByPath(path)
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
		case err != nil:
			return fmt.Errorf("db error: %w", err)
		}
		return nil
	}

	objectID, err := resolveFileID(path)
	if err != nil {
		return fmt.Errorf("resolve ObjectID: %w", err)
	}

	err = indexer.mod.db.UpsertPath(path, objectID, stat.ModTime())
	if err != nil {
		return fmt.Errorf("db upsert: %w", err)
	}

	if indexer.subscriberCount.Load() > 0 {
		indexer.events = indexer.events.Push(IndexEvent{
			Path:     path,
			ObjectID: objectID,
		})
	}

	return nil
}

func (indexer *Indexer) init(ctx *astral.Context) error {
	// todo: roots can overlap (we could reduce it to a set of wide roots)

	// discover new files
	for _, root := range indexer.roots.Clone() {
		err := indexer.scan(root)
		if err != nil {
			return err
		}
	}

	// invalidate all steps
	return indexer.mod.db.EachPath(func(s string) error {
		return indexer.invalidate(s)
	})
}

func (indexer *Indexer) scan(root string) error {
	return filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !entry.Type().IsRegular() {
			return nil
		}

		err = indexer.invalidate(path)
		if err != nil {
			return err
		}

		return nil
	})
}

func (indexer *Indexer) addRoot(root string) error {
	return indexer.roots.Add(root)
}

func (indexer *Indexer) removeRoot(root string) (err error) {
	return indexer.roots.Remove(root)
}

func (indexer *Indexer) subscribe() *sig.Queue[IndexEvent] {
	indexer.subscriberCount.Add(1)
	return indexer.events
}

func (indexer *Indexer) unsubscribe() {
	indexer.subscriberCount.Add(-1)
}
