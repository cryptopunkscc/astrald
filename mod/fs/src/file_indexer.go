package fs

import (
	"sync"

	"github.com/cryptopunkscc/astrald/astral"
)

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
type FileIndexer interface {
	// Acquire registers interest in a path. Must be called before MarkDirty.
	// Multiple calls from different repositories increment the interest count.
	Acquire(path string)
	// Release removes interest in a path. Call when a repository closes.
	// If interest drops to 0, pending indexing for this path will be skipped.
	Release(path string)

	// MarkDirty queues a path for indexing. No-op if interest == 0.
	MarkDirty(path string)

	Close() error
	Subscribe(ctx *astral.Context) <-chan *astral.ObjectID
}

type updateState struct {
	queued   bool
	running  bool
	rerun    bool
	interest int
}

type fileIndexer struct {
	mod *Module

	work chan string
	done chan struct{}

	mu     sync.Mutex
	states map[string]*updateState

	wg   sync.WaitGroup
	once sync.Once
}

func NewFileIndexer(mod *Module, workers int, queueLen int) FileIndexer {
	if workers <= 0 {
		workers = 1
	}
	if queueLen <= 0 {
		queueLen = 1
	}

	fi := &fileIndexer{
		mod:    mod,
		work:   make(chan string, queueLen),
		done:   make(chan struct{}),
		states: map[string]*updateState{},
	}

	fi.wg.Add(workers)
	for i := 0; i < workers; i++ {
		go fi.worker()
	}

	return fi
}

func (fi *fileIndexer) Acquire(path string) {
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

	st := fi.states[path]
	if st == nil {
		st = &updateState{}
		fi.states[path] = st
	}
	st.interest++
}

func (fi *fileIndexer) Release(path string) {
	if len(path) == 0 || path[0] != '/' {
		return
	}

	fi.mu.Lock()
	defer fi.mu.Unlock()

	st := fi.states[path]
	if st == nil {
		return
	}

	st.interest--
	if st.interest < 0 {
		st.interest = 0
	}

	// If no interest and not running/queued, clean up state
	if st.interest == 0 && !st.running && !st.queued {
		delete(fi.states, path)
	}
}

func (fi *fileIndexer) MarkDirty(path string) {
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

	st := fi.states[path]
	// No-op if no state exists or interest is 0
	if st == nil || st.interest == 0 {
		fi.mu.Unlock()
		return
	}

	// If processing is currently running, remember we need to rerun after it completes.
	if st.running {
		st.rerun = true
		fi.mu.Unlock()
		return
	}

	// If already queued, nothing to do.
	if st.queued {
		fi.mu.Unlock()
		return
	}

	st.queued = true
	fi.mu.Unlock()

	// Enqueue outside the lock using select to handle shutdown.
	select {
	case <-fi.done:
		return
	case fi.work <- path:
	}
}

func (fi *fileIndexer) Close() error {
	fi.once.Do(func() {
		close(fi.done)
		fi.wg.Wait()
	})
	return nil
}

func (fi *fileIndexer) worker() {
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

func (fi *fileIndexer) processPath(path string) {
	fi.mu.Lock()
	st := fi.states[path]
	if st == nil {
		fi.mu.Unlock()
		return
	}

	// Skip if no interest remains
	if st.interest == 0 {
		st.queued = false
		st.running = false
		st.rerun = false
		delete(fi.states, path)
		fi.mu.Unlock()
		return
	}

	st.queued = false
	st.running = true
	fi.mu.Unlock()

	objectID, err := fi.mod.updateDbIndex(path)
	if err == nil && objectID != nil {
		fi.mod.pushFileIndexed(objectID)
	}

	fi.mu.Lock()
	st = fi.states[path]
	if st == nil {
		fi.mu.Unlock()
		return
	}
	st.running = false

	// Only rerun if interest remains and rerun was requested
	if st.rerun && st.interest > 0 {
		st.rerun = false
		st.queued = true
		fi.mu.Unlock()

		// Re-enqueue using select to handle shutdown
		select {
		case <-fi.done:
			return
		case fi.work <- path:
		}
		return
	}

	// Clear rerun flag (we're not going to rerun)
	st.rerun = false

	// Only delete state if no interest remains.
	// If interest > 0, keep state so future MarkDirty calls work.
	if st.interest == 0 {
		delete(fi.states, path)
	}
	fi.mu.Unlock()
}

func (fi *fileIndexer) Subscribe(ctx *astral.Context) <-chan *astral.ObjectID {
	return fi.mod.subscribeFileIndexed(ctx)
}
