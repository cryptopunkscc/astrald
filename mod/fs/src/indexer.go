package fs

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/lib/paths"
	"github.com/cryptopunkscc/astrald/mod/fs"
	"github.com/cryptopunkscc/astrald/sig"
	"golang.org/x/time/rate"
	"gorm.io/gorm"
)

const (
	statRate  = 200
	statBurst = 800
	hashRate  = 30
	hashBurst = 150
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

	mu              sync.Mutex
	events          *sig.Queue[IndexEvent]
	subscriberCount int

	// rate limitters to prevent too many I/O operations
	statLimiter *rate.Limiter
	hashLimiter *rate.Limiter
}

func NewIndexer(mod *Module) *Indexer {
	indexer := &Indexer{
		mod:       mod,
		log:       mod.log,
		workqueue: make(chan string, 1024),
		events:    &sig.Queue[IndexEvent]{},

		statLimiter: rate.NewLimiter(rate.Limit(statRate), statBurst),
		hashLimiter: rate.NewLimiter(rate.Limit(hashRate), hashBurst),
	}

	return indexer
}

func (indexer *Indexer) startWorkers(ctx *astral.Context, count int) {
	for i := 0; i < count; i++ {
		go indexer.worker(ctx)
	}
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
				if paths.PathUnderRoot(path, root) {
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

			if err := indexer.checkAndFix(ctx, path); err != nil {
				indexer.log.Error("indexer: checkAndFix %v: %v", path, err)
			}
		}
	}

}

func (indexer *Indexer) remove(path string) (err error) {
	err = indexer.mod.db.DeleteByPath(path)
	if err != nil {
		return fmt.Errorf("db delete: %w", err)
	}

	return nil
}

func (indexer *Indexer) enqueue(path string) {
	if err := indexer.pending.Add(path); err == nil {
		indexer.workqueue <- path
	}
}

func (indexer *Indexer) invalidate(path string) (err error) {
	err = indexer.mod.db.Invalidate(path)
	if err != nil {
		return fmt.Errorf("db invalidate: %w", err)
	}

	indexer.enqueue(path)
	return
}

func (indexer *Indexer) checkAndFix(ctx context.Context, path string) error {
	if len(path) == 0 || path[0] != '/' {
		return fs.ErrInvalidPath
	}

	if err := indexer.statLimiter.Wait(ctx); err != nil {
		return err
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

	indexEntry, err := indexer.mod.db.FindByPath(path)
	if err == nil && indexEntry.ModTime == stat.ModTime() {
		return indexer.mod.db.ValidatePath(path)
	}

	if err := indexer.hashLimiter.Wait(ctx); err != nil {
		return err
	}

	objectID, err := resolveFileID(path)
	if err != nil {
		return fmt.Errorf("resolve ObjectID: %w", err)
	}

	err = indexer.mod.db.UpsertPath(path, objectID, stat.ModTime())
	if err != nil {
		return fmt.Errorf("db upsert: %w", err)
	}

	indexer.mu.Lock()
	if indexer.subscriberCount > 0 {
		indexer.events = indexer.events.Push(IndexEvent{
			Path:     path,
			ObjectID: objectID,
		})
	}
	indexer.mu.Unlock()

	return nil
}

func (indexer *Indexer) init(ctx *astral.Context) error {
	now := time.Now()

	// discover new files
	for _, root := range paths.WidestRoots(indexer.roots.Clone()) {
		if err := indexer.scan(ctx, root); err != nil {
			return err
		}
	}
	indexer.mod.log.Info(`fs indexer: scan completed in %v`, time.Since(now))
	if err := indexer.mod.db.InvalidateAllPaths(); err != nil {
		return fmt.Errorf("invalidate all paths: %w", err)
	}

	return indexer.mod.db.EachInvalidPath(func(s string) error {
		indexer.enqueue(s)
		return nil
	})
}

func (indexer *Indexer) scan(ctx context.Context, root string) error {
	return filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !entry.Type().IsRegular() {
			return nil
		}

		err = indexer.statLimiter.Wait(ctx)
		if err != nil {
			return err
		}

		err = indexer.invalidate(path)
		if err != nil {
			return err
		}

		return nil
	})
}

func (indexer *Indexer) addRoot(root string) error {
	return indexer.roots.Add(filepath.Clean(root))
}

func (indexer *Indexer) removeRoot(root string) error {
	root = filepath.Clean(root)
	roots := indexer.roots.Clone()

	// no deletion needed (still covered by other roots)
	for _, other := range roots {
		if other != root && paths.PathUnderRoot(root, other) {
			return indexer.roots.Remove(root)
		}
	}

	// Find narrower roots under the root being removed
	var narrower []string
	for _, other := range roots {
		if other != root && paths.PathUnderRoot(other, root) {
			narrower = append(narrower, other)
		}
	}

	// Build trie from narrower roots for O(L) coverage check
	trie, err := paths.NewPathTrie(narrower)
	if err != nil {
		if errors.Is(err, paths.ErrNotAbsolute) {
			return fs.ErrNotAbsolute
		}
		return fmt.Errorf("build trie: %w", err)
	}

	// Delete paths not covered by narrower roots
	err = indexer.mod.db.EachPath(root, func(path string) error {
		covered, err := trie.Covers(path)
		if err != nil {
			if errors.Is(err, paths.ErrNotAbsolute) {
				return fs.ErrNotAbsolute
			}
			return err
		}
		if !covered {
			return indexer.mod.db.DeleteByPath(path)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("delete paths: %w", err)
	}

	return indexer.roots.Remove(root)
}

func (indexer *Indexer) subscribe() *sig.Queue[IndexEvent] {
	indexer.mu.Lock()
	defer indexer.mu.Unlock()
	indexer.subscriberCount++
	return indexer.events
}

func (indexer *Indexer) unsubscribe() {
	indexer.mu.Lock()
	defer indexer.mu.Unlock()
	indexer.subscriberCount--
}
