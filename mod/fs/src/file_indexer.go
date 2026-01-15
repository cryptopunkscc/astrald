package fs

import (
	"context"
	"strings"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/sig"
)

// FileIndexer is responsible for indexing files in the filesystem, it only coordinates the work of indexing files.
// Core semantics/invariants:
// - Indexing is done by workers (goroutines)
// - MarkDirty is method called frequently by many producers (like WatchRepository) to schedule indexing of files due to fsnotjfy changes
// - If we mark path dirty while its already scheduled for indexing, we dont need to add it again, it will be picked up by the worker. in its time
// - When repository is removed via ReleaseRoot, queued paths for that root become invalidated but remain in queue
// - Workers check if path is still under an active root before indexing (fast check, lazy cleanup)
// - MarkDirty immediately deletes the DB entry to prevent stale reads while file is waiting to be re-indexed
type FileIndexer struct {
	mod       *Module
	workqueue *sig.DedupQueue[string] // paths to index waiting in queue (deduplicated)

	ctx    context.Context
	cancel context.CancelFunc

	activeRoots sig.Set[string] // roots currently being watched by repositories

	// After indexing published updates
	updates *sig.Queue[*astral.ObjectID]
}

// NewFileIndexer creates a new FileIndexer with the specified number of workers.
func NewFileIndexer(mod *Module, workers int) *FileIndexer {
	ctx, cancel := context.WithCancel(context.Background())

	fi := &FileIndexer{
		mod:         mod,
		workqueue:   sig.NewDedupQueue[string](),
		ctx:         ctx,
		cancel:      cancel,
		activeRoots: sig.Set[string]{},
		updates:     &sig.Queue[*astral.ObjectID]{},
	}

	for i := 0; i < workers; i++ {
		go fi.worker()
	}

	return fi
}

// worker pulls paths from the workqueue and indexes them.
func (fi *FileIndexer) worker() {
	for {
		path, ok := fi.workqueue.Dequeue()
		if !ok {
			return
		}

		// Check cancellation before starting expensive work
		select {
		case <-fi.ctx.Done():
			return
		default:
		}

		// Fast check: skip if path is no longer under any active root (lazy cleanup)
		if !fi.IsUnderActiveRoot(path) {
			continue
		}

		objectID, err := fi.mod.updateDbIndex(path)
		if err != nil {
			continue
		}

		if objectID != nil {
			fi.updates.Push(objectID)
		}
	}
}

// IsUnderActiveRoot checks if path is under any active root.
// Properly handles directory boundaries to avoid false matches like:
// "/workspace/product-tools" being matched by root "/workspace/product"
func (fi *FileIndexer) IsUnderActiveRoot(path string) bool {
	for _, root := range fi.activeRoots.Clone() {
		if path == root || strings.HasPrefix(path, root+"/") {
			return true
		}
	}
	return false
}

// AcquireRoot registers in a root path.
func (fi *FileIndexer) AcquireRoot(root string) {
	fi.activeRoots.Add(root)
}

// ReleaseRoot unregisters interest in a root path.
// Paths under this root will be skipped during indexing (if they are not tracked by wider root).
func (fi *FileIndexer) ReleaseRoot(root string) {
	fi.activeRoots.Remove(root)
}

// CleanupRoot removes DB entries for paths under the given root that are no longer tracked.
func (fi *FileIndexer) CleanupRoot(root string) (deleted int, err error) {
	err = fi.mod.db.EachPathWithPrefix(root, func(path string) error {
		if !fi.IsUnderActiveRoot(path) {
			// Path is orphaned - delete from DB
			_ = fi.mod.db.DeleteByPath(path)
			deleted++
		}
		return nil
	})
	return deleted, err
}

// MarkDirty marks the given path as dirty and schedules it for indexing.
func (fi *FileIndexer) MarkDirty(path string) {
	// Delete stale DB entry immediately to prevent returning wrong content
	// If the file doesn't exist or isn't in DB, this is a no-op
	_ = fi.mod.db.DeleteByPath(path)

	// Enqueue for re-indexing
	fi.workqueue.Enqueue(path)
}

// Subscribe returns a channel that receives ObjectIDs as files are indexed.
func (fi *FileIndexer) Subscribe(ctx context.Context) <-chan *astral.ObjectID {
	return fi.updates.Subscribe(ctx)
}

// Close gracefully shuts down the FileIndexer.
func (fi *FileIndexer) Close() error {
	fi.cancel()
	fi.workqueue.Close()
	fi.updates.Close()
	return nil
}
