package fs

import (
	"fmt"
	"sync"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/sig"
)

type IndexFunc func(path string) (*astral.ObjectID, error)

// FileIndexer coordinates indexing work for absolute filesystem paths with root-scoped interest.
// All methods are safe to call from multiple goroutines.
// Close stops workers and waits for them to exit.
// Subscribe returns object IDs produced by successful indexing.
type FileIndexer struct {
	indexFn IndexFunc

	// Worker coordination
	pathQueue *sig.WorkQueue[string] // global path queue (workers read from this)
	updates   *sig.Queue[*astral.ObjectID]

	// Hierarchy tracking
	rootTree *RootTree

	// State
	mu     sync.Mutex
	closed bool
	roots  map[string]*rootEntry        // root → interest tracking
	files  map[string]*fileIndexerEntry // path → execution state (transient)

	wg   sync.WaitGroup
	once sync.Once
}

// rootEntry tracks per-root interest
type rootEntry struct {
	interest int // reference count from repositories
}

// NewFileIndexer creates a new FileIndexer with the given indexing function.
// indexFn is called for each path that needs indexing; it must be non-nil.
// workers controls concurrency (min 1).
// pathQueueLen and rootQueueLen parameters are ignored.
func NewFileIndexer(indexFn IndexFunc, workers int, pathQueueLen int, rootQueueLen int) (*FileIndexer, error) {
	if indexFn == nil {
		return nil, fmt.Errorf("indexFn cannot be nil")
	}

	if workers <= 0 {
		workers = 1
	}

	fi := &FileIndexer{
		indexFn:   indexFn,
		rootTree:  NewRootTree(),
		pathQueue: sig.NewWorkQueue[string](),
		updates:   &sig.Queue[*astral.ObjectID]{},
		roots:     make(map[string]*rootEntry),
		files:     make(map[string]*fileIndexerEntry),
	}

	// Start path workers
	fi.wg.Add(workers)
	for i := 0; i < workers; i++ {
		go fi.pathWorker()
	}

	return fi, nil
}

// AcquireRoot registers interest in a repository root.
// All files under this root are implicitly interesting.
// This is an O(1) operation - no file iteration.
func (fi *FileIndexer) AcquireRoot(root string) {
	if len(root) == 0 || root[0] != '/' {
		return
	}

	fi.mu.Lock()
	defer fi.mu.Unlock()

	if fi.closed {
		return
	}

	// Register in hierarchy
	fi.rootTree.Add(root)

	e := fi.roots[root]
	if e == nil {
		e = &rootEntry{}
		fi.roots[root] = e
	}
	e.interest++

	// No file iteration - lazy discovery on next MarkDirty
}

// ReleaseRoot decrements interest in a root.
// When interest drops to 0, all file state under this root is cleaned up.
func (fi *FileIndexer) ReleaseRoot(root string) {
	if len(root) == 0 || root[0] != '/' {
		return
	}

	fi.mu.Lock()
	defer fi.mu.Unlock()

	e := fi.roots[root]
	if e == nil {
		return
	}

	e.interest--
	if e.interest <= 0 {
		// Remove from hierarchy
		fi.rootTree.Remove(root)

		// Clean up file entries that reference this root
		for path, fileEntry := range fi.files {
			delete(fileEntry.roots, root)

			// If file has no more watching roots, delete it
			if len(fileEntry.roots) == 0 && fileEntry.canDelete() {
				delete(fi.files, path)
			}
		}

		delete(fi.roots, root)
	}
}

// MarkDirty marks a file path as needing indexing.
// Uses lazy root discovery and global dedup via state machine.
func (fi *FileIndexer) MarkDirty(path string) {
	if len(path) == 0 || path[0] != '/' {
		return
	}

	fi.mu.Lock()

	if fi.closed {
		fi.mu.Unlock()
		return
	}

	// Find all roots containing this path
	watchingRoots := fi.findRootsContaining(path)
	if len(watchingRoots) == 0 {
		fi.mu.Unlock()
		return
	}

	// Get or create file entry
	fileEntry := fi.files[path]
	if fileEntry == nil {
		fileEntry = &fileIndexerEntry{
			roots: make(map[string]struct{}),
		}
		fi.files[path] = fileEntry
	}

	// Update root associations (lazy discovery)
	fileEntry.roots = make(map[string]struct{})
	for _, root := range watchingRoots {
		fileEntry.roots[root] = struct{}{}
	}

	// Check if any watching root has interest
	if !fileEntry.hasInterest(fi) {
		fi.mu.Unlock()
		return
	}

	// Already queued or running? State machine handles it
	if fileEntry.queued || fileEntry.running {
		if fileEntry.running {
			fileEntry.rerun = true
		}
		fi.mu.Unlock()
		return
	}

	// Mark as queued and enqueue directly
	fileEntry.queued = true
	fi.mu.Unlock()

	fi.pathQueue.Enqueue(path)
}

// findRootsContaining returns all roots that contain path.
// Must be called with fi.mu held.
func (fi *FileIndexer) findRootsContaining(path string) []string {
	// Use RootTree for hierarchy-aware discovery
	allRoots := fi.rootTree.FindAll(path)

	// Filter by interest
	var result []string
	for _, root := range allRoots {
		if entry := fi.roots[root]; entry != nil && entry.interest > 0 {
			result = append(result, root)
		}
	}
	return result
}

func (fi *FileIndexer) Close() error {
	fi.once.Do(func() {
		// Mark as closed
		fi.mu.Lock()
		fi.closed = true
		fi.mu.Unlock()

		// Close queue (wakes all blocked goroutines)
		fi.pathQueue.Close()

		// Wait for all workers to exit
		fi.wg.Wait()

		// Close updates queue
		fi.updates.Close()
	})
	return nil
}

// pathWorker pulls paths from global queue and processes them
func (fi *FileIndexer) pathWorker() {
	defer fi.wg.Done()

	for {
		path, ok := fi.pathQueue.Dequeue()
		if !ok {
			return // queue closed and drained
		}

		fi.processPath(path)
	}
}

func (fi *FileIndexer) processPath(path string) {
	fi.mu.Lock()

	fileEntry := fi.files[path]
	if fileEntry == nil || !fileEntry.claim() {
		if fileEntry != nil && fileEntry.canDelete() {
			delete(fi.files, path)
		}
		fi.mu.Unlock()
		return
	}

	// Claim succeeded - check interest one more time
	if !fileEntry.hasInterest(fi) {
		fileEntry.reset()
		if fileEntry.canDelete() {
			delete(fi.files, path)
		}
		fi.mu.Unlock()
		return
	}

	fi.mu.Unlock()

	// Run indexing function (outside mutex)
	objectID, err := fi.indexFn(path)
	if err == nil && objectID != nil {
		fi.updates = fi.updates.Push(objectID)
	}

	fi.mu.Lock()

	fileEntry = fi.files[path]
	if fileEntry == nil {
		fi.mu.Unlock()
		return
	}

	needsRequeue := fileEntry.indexed()
	if needsRequeue {
		// Mark dirty again (will go through per-root dedup)
		fi.mu.Unlock()
		fi.MarkDirty(path)
		return
	}

	if fileEntry.canDelete() {
		delete(fi.files, path)
	}

	fi.mu.Unlock()
}

func (fi *FileIndexer) Subscribe(ctx *astral.Context) <-chan *astral.ObjectID {
	if ctx == nil {
		ctx = astral.NewContext(nil)
	}
	return fi.updates.Subscribe(ctx)
}

// fileIndexerEntry is a state machine for a single path's indexing state.
// States: idle -> queued -> running -> idle (or back to queued if rerun needed)
type fileIndexerEntry struct {
	roots   map[string]struct{} // which roots watch this file
	queued  bool
	running bool
	rerun   bool
}

// hasInterest returns true if any root watching this file has interest > 0
func (e *fileIndexerEntry) hasInterest(fi *FileIndexer) bool {
	for root := range e.roots {
		if rootEntry := fi.roots[root]; rootEntry != nil && rootEntry.interest > 0 {
			return true
		}
	}
	return false
}

// claim attempts to claim this entry for processing.
// Returns true if indexing should proceed, false to skip (no roots watching).
func (e *fileIndexerEntry) claim() bool {
	if len(e.roots) == 0 {
		e.reset()
		return false
	}
	e.queued = false
	e.running = true
	return true
}

// indexed marks indexing as complete.
// Returns true if re-enqueue is needed (dirty while running, roots remain).
func (e *fileIndexerEntry) indexed() bool {
	e.running = false
	if e.rerun && len(e.roots) > 0 {
		e.rerun = false
		return true
	}
	e.rerun = false
	return false
}

// canDelete returns true if this entry can be removed from the map.
func (e *fileIndexerEntry) canDelete() bool {
	return len(e.roots) == 0 && !e.running && !e.queued
}

// reset clears all flags (used when skipping due to no interest).
func (e *fileIndexerEntry) reset() {
	e.queued = false
	e.running = false
	e.rerun = false
}

// TopLevelRoots returns the widest roots being tracked.
func (fi *FileIndexer) TopLevelRoots() []string {
	fi.mu.Lock()
	defer fi.mu.Unlock()
	return fi.rootTree.TopLevel()
}

// FindWidestRoot returns the widest root containing path.
func (fi *FileIndexer) FindWidestRoot(path string) string {
	fi.mu.Lock()
	defer fi.mu.Unlock()
	return fi.rootTree.FindWidest(path)
}
