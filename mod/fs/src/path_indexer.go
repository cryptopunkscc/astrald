package fs

import (
	"sync"

	"github.com/cryptopunkscc/astrald/astral"
)

// PathIndexer coordinates indexing work for absolute filesystem paths.
//
// Callers report that a path is "dirty" rather than invoking indexing directly.
// The PathIndexer deduplicates requests per (owner,path) and ensures that at most one
// indexing operation for a given path runs at a time.
//
// If a path is marked dirty while it is currently being indexed, PathIndexer guarantees
// that the path will be indexed again after the in-progress indexing completes.
//
// The actual indexing logic is delegated to Module.updateDbIndex(path).
//
// MarkDirty/MarkDirtyOwned are safe to call from multiple goroutines.
// Close stops workers and waits for them to exit.
// DropOwner removes queued and rerun work for the owner.
//
// Owner is typically the repository label that originated the update.
// It allows removing a repo to also remove its pending indexing work.
//
// Note: DropOwner does not (and cannot safely) stop an in-flight updateDbIndex call,
// but it will prevent reruns and future queued work for that owner.
//
// Paths are required to be absolute.
//
// Subscribe returns object IDs produced by successful indexing.
// The stream is global; callers can filter by root path using the DB if needed.
//
// Backwards-compatibility: MarkDirty(path) behaves like MarkDirtyOwned("", path).
//
// Close is idempotent.
//
// DropOwner is idempotent.
type PathIndexer interface {
	MarkDirty(path string)
	MarkDirtyOwned(owner string, path string)

	// DropOwner drops all queued and rerun work for the given owner.
	DropOwner(owner string)

	Close() error

	Subscribe(ctx *astral.Context) <-chan *astral.ObjectID
}

type pathUpdateState struct {
	queued  bool
	running bool
	rerun   bool

	owner string
}

type pathIndexer struct {
	mod *Module

	work chan string

	mu     sync.Mutex
	states map[string]*pathUpdateState
	closed bool

	wg   sync.WaitGroup
	once sync.Once
}

func NewPathIndexer(mod *Module, workers int, queueLen int) PathIndexer {
	if workers <= 0 {
		workers = 1
	}
	if queueLen <= 0 {
		queueLen = 1
	}

	pi := &pathIndexer{
		mod:    mod,
		work:   make(chan string, queueLen),
		states: map[string]*pathUpdateState{},
	}

	pi.wg.Add(workers)
	for i := 0; i < workers; i++ {
		go pi.worker()
	}

	return pi
}

func (i *pathIndexer) MarkDirty(path string) {
	i.MarkDirtyOwned("", path)
}

func (i *pathIndexer) MarkDirtyOwned(owner string, path string) {
	if len(path) == 0 || path[0] != '/' {
		return
	}

	i.mu.Lock()
	if i.closed {
		i.mu.Unlock()
		return
	}

	st := i.states[path]
	if st == nil {
		st = &pathUpdateState{owner: owner}
		i.states[path] = st
	}

	// If state exists but is owned by a different owner, do not merge.
	// The first owner to queue the path "wins" until it is drained.
	if st.owner != owner {
		i.mu.Unlock()
		return
	}

	// If processing is currently running, remember we need to rerun after it completes.
	if st.running {
		st.rerun = true
		i.mu.Unlock()
		return
	}

	// If already queued, nothing to do.
	if st.queued {
		i.mu.Unlock()
		return
	}

	st.queued = true
	i.mu.Unlock()

	// Enqueue outside the lock.
	// If Close() raced and the channel is closed, recover and drop.
	defer func() { _ = recover() }()
	i.work <- path
}

func (i *pathIndexer) DropOwner(owner string) {
	i.mu.Lock()
	defer i.mu.Unlock()

	for path, st := range i.states {
		if st == nil {
			continue
		}
		if st.owner != owner {
			continue
		}

		// If queued but not running, remove entirely.
		if st.queued && !st.running {
			delete(i.states, path)
			continue
		}

		// If running, prevent rerun.
		if st.running {
			st.rerun = false
		}
	}
}

func (i *pathIndexer) Close() error {
	i.once.Do(func() {
		i.mu.Lock()
		i.closed = true
		close(i.work)
		i.mu.Unlock()

		i.wg.Wait()
	})
	return nil
}

func (i *pathIndexer) worker() {
	defer i.wg.Done()

	for path := range i.work {
		i.mu.Lock()
		st := i.states[path]
		if st == nil {
			// Work item could be stale (e.g. owner dropped). Skip.
			i.mu.Unlock()
			continue
		}
		st.queued = false
		st.running = true
		owner := st.owner
		i.mu.Unlock()

		objectID, err := i.mod.updateDbIndex(path)
		if err == nil && objectID != nil {
			i.mod.pushPathIndexed(objectID)
		}

		i.mu.Lock()
		st = i.states[path]
		if st == nil {
			// Dropped while running: nothing else to do.
			i.mu.Unlock()
			continue
		}
		st.running = false

		// If owner changed/dropped, do not reschedule.
		if st.owner != owner {
			delete(i.states, path)
			i.mu.Unlock()
			continue
		}

		if st.rerun {
			st.rerun = false
			st.queued = true
			i.mu.Unlock()
			// re-enqueue (guard against Close races)
			func() {
				defer func() { _ = recover() }()
				i.work <- path
			}()
			continue
		}

		delete(i.states, path)
		i.mu.Unlock()
	}
}

func (i *pathIndexer) Subscribe(ctx *astral.Context) <-chan *astral.ObjectID {
	return i.mod.subscribePathIndexed(ctx)
}
