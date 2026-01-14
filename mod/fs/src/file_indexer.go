package fs

import (
	"sync"

	"github.com/cryptopunkscc/astrald/astral"
)

// fileIndexerEntry is a state machine for a single path's indexing state.
// States: idle -> queued -> running -> idle (or back to queued if rerun needed)
type fileIndexerEntry struct {
	queued   bool
	running  bool
	rerun    bool
	interest int
}

// setDirty marks this entry as needing indexing.
// Returns true if the caller should enqueue to the work channel.
func (e *fileIndexerEntry) setDirty() bool {
	if e.interest == 0 {
		return false
	}
	if e.running {
		e.rerun = true
		return false
	}
	if e.queued {
		return false
	}
	e.queued = true
	return true
}

// claim attempts to claim this entry for processing.
// Returns true if indexing should proceed, false to skip (no interest).
func (e *fileIndexerEntry) claim() bool {
	if e.interest == 0 {
		e.reset()
		return false
	}
	e.queued = false
	e.running = true
	return true
}

// indexed marks indexing as complete.
// Returns true if re-enqueue is needed (dirty while running, interest remains).
func (e *fileIndexerEntry) indexed() bool {
	e.running = false
	if e.rerun && e.interest > 0 {
		e.rerun = false
		e.queued = true
		return true
	}
	e.rerun = false
	return false
}

// canDelete returns true if this entry can be removed from the map.
func (e *fileIndexerEntry) canDelete() bool {
	return e.interest == 0 && !e.running && !e.queued
}

// reset clears all flags (used when skipping due to no interest).
func (e *fileIndexerEntry) reset() {
	e.queued = false
	e.running = false
	e.rerun = false
}

// FileIndexer coordinates indexing work for absolute filesystem paths.
//
// Callers report that a path is "dirty" based on their observation (e.g file system events).
// The FileIndexer deduplicates requests per path and ensures that at most one
// indexing operation for a given path runs at a time.
//
// If a path is marked dirty while it is currently being indexed, FileIndexer guarantees
// that the path will be indexed again after the in-progress indexing completes.
//
// Interest tracking:
// Repositories must call Acquire(path) before MarkDirty(path) to register interest.
// When a repository closes, it calls Release(path) for all acquired paths.
// If interest drops to 0, pending/in-flight indexing for that path is skipped.
//
// All methods are safe to call from multiple goroutines.
// Close stops workers and waits for them to exit.
// Subscribe returns object IDs produced by successful indexing.
type FileIndexer struct {
	mod *Module

	work chan string
	done chan struct{}

	mu     sync.Mutex
	states map[string]*fileIndexerEntry

	wg   sync.WaitGroup
	once sync.Once
}

func NewFileIndexer(mod *Module, workers int, queueLen int) *FileIndexer {
	if workers <= 0 {
		workers = 1
	}
	if queueLen <= 0 {
		queueLen = 1
	}

	fi := &FileIndexer{
		mod:    mod,
		work:   make(chan string, queueLen),
		done:   make(chan struct{}),
		states: map[string]*fileIndexerEntry{},
	}

	fi.wg.Add(workers)
	for i := 0; i < workers; i++ {
		go fi.worker()
	}

	return fi
}

func (fi *FileIndexer) Acquire(path string) {
	if len(path) == 0 || path[0] != '/' {
		return
	}

	fi.mu.Lock()
	defer fi.mu.Unlock()

	select {
	case <-fi.done:
		return
	default:
	}

	e := fi.states[path]
	if e == nil {
		e = &fileIndexerEntry{}
		fi.states[path] = e
	}
	e.interest++
}

func (fi *FileIndexer) Release(path string) {
	if len(path) == 0 || path[0] != '/' {
		return
	}

	fi.mu.Lock()
	defer fi.mu.Unlock()

	e := fi.states[path]
	if e == nil {
		return
	}

	e.interest--
	if e.interest < 0 {
		e.interest = 0
	}

	if e.canDelete() {
		delete(fi.states, path)
	}
}

func (fi *FileIndexer) MarkDirty(path string) {
	if len(path) == 0 || path[0] != '/' {
		return
	}

	fi.mu.Lock()

	select {
	case <-fi.done:
		fi.mu.Unlock()
		return
	default:
	}

	e := fi.states[path]
	if e == nil || !e.setDirty() {
		fi.mu.Unlock()
		return
	}
	fi.mu.Unlock()

	select {
	case <-fi.done:
	case fi.work <- path:
	}
}

func (fi *FileIndexer) Close() error {
	fi.once.Do(func() {
		close(fi.done)
		fi.wg.Wait()
	})
	return nil
}

func (fi *FileIndexer) worker() {
	defer fi.wg.Done()

	for {
		select {
		case <-fi.done:
			return
		case path := <-fi.work:
			fi.processPath(path)
		}
	}
}

func (fi *FileIndexer) processPath(path string) {
	fi.mu.Lock()
	e := fi.states[path]
	if e == nil || !e.claim() {
		if e != nil && e.canDelete() {
			delete(fi.states, path)
		}
		fi.mu.Unlock()
		return
	}
	fi.mu.Unlock()

	objectID, err := fi.mod.updateDbIndex(path)
	if err == nil && objectID != nil {
		fi.mod.pushFileIndexed(objectID)
	}

	fi.mu.Lock()
	e = fi.states[path]
	if e == nil {
		fi.mu.Unlock()
		return
	}

	needsRequeue := e.indexed()
	if e.canDelete() {
		delete(fi.states, path)
	}
	fi.mu.Unlock()

	if needsRequeue {
		select {
		case <-fi.done:
		case fi.work <- path:
		}
	}
}

func (fi *FileIndexer) Subscribe(ctx *astral.Context) <-chan *astral.ObjectID {
	return fi.mod.subscribeFileIndexed(ctx)
}
