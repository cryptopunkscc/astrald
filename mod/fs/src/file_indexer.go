package fs

import (
	"context"
	"strings"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/sig"
)

type IndexFunc func(path string) (*astral.ObjectID, error)

// FileIndexer is responsible for indexing files in the filesystem, it only coordinates the work of indexing files.
// Core semantics/invariants:
// - Indexing is done by workers (goroutines)
// - MarkDirty is method called frequently by many producers (like WatchRepository) to schedule indexing of files due to fsnotjfy changes
// - If we mark path dirty while its already scheduled for indexing, we dont need to add it again, it will be picked up by the worker. in its time
// - When repository is removed via ReleaseRoot, queued paths for that root become invalidated but remain in queue
// - Workers check if path is still under an active root before indexing (fast check, lazy cleanup)
type FileIndexer struct {
	indexFn   IndexFunc
	workqueue *sig.DedupQueue[string] // paths to index waiting in queue (deduplicated)
	closed    bool

	activeRoots sig.Set[string] // roots currently being watched by repositories

	// After indexing published updates
	updates *sig.Queue[*astral.ObjectID]
}

// NewFileIndexer creates a new FileIndexer with the specified number of workers.
func NewFileIndexer(indexFn IndexFunc, workers int) *FileIndexer {
	fi := &FileIndexer{
		indexFn:     indexFn,
		workqueue:   sig.NewDedupQueue[string](),
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

		// Fast check: skip if path is no longer under any active root (lazy cleanup)
		if !fi.isUnderActiveRoot(path) {
			continue
		}

		objectID, err := fi.indexFn(path)
		if err != nil {
			continue
		}

		if objectID != nil {
			fi.updates.Push(objectID)
		}
	}
}

// isUnderActiveRoot checks if path is under any active root.
// Properly handles directory boundaries to avoid false matches like:
// "/workspace/product-tools" being matched by root "/workspace/product"
func (fi *FileIndexer) isUnderActiveRoot(path string) bool {
	for _, root := range fi.activeRoots.Clone() {
		if path == root || strings.HasPrefix(path, root+"/") {
			return true
		}
	}
	return false
}

// AcquireRoot registers interest in a root path.
func (fi *FileIndexer) AcquireRoot(root string) {
	fi.activeRoots.Add(root)
}

// ReleaseRoot unregisters interest in a root path.
// Paths under this root will be skipped during indexing.
func (fi *FileIndexer) ReleaseRoot(root string) {
	fi.activeRoots.Remove(root)
}

// MarkDirty marks the given path as dirty and schedules it for indexing.
// This is a fast O(1) operation that automatically deduplicates - if the path
// is already queued for indexing, it won't be added again.
func (fi *FileIndexer) MarkDirty(path string) error {
	fi.workqueue.Enqueue(path)
	return nil
}

// Subscribe returns a channel that receives ObjectIDs as files are indexed.
// The channel will be closed when the context is canceled.
// Multiple subscribers can listen concurrently.
func (fi *FileIndexer) Subscribe(ctx context.Context) <-chan *astral.ObjectID {
	return fi.updates.Subscribe(ctx)
}

func (fi *FileIndexer) Close() error {
	// TODO: implement later
	return nil
}
