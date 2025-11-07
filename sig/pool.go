package sig

import "sync"

type Pool struct {
	cap    map[string]int
	locked map[string]int
	cond   *sync.Cond
}

func NewPool() *Pool {
	return &Pool{
		cap:    map[string]int{},
		locked: map[string]int{},
		cond:   sync.NewCond(&sync.Mutex{}),
	}
}

// Add adds an amount of an item to the pool
func (r *Pool) Add(item string, count int) int {
	r.cond.L.Lock()
	defer r.cond.L.Unlock()

	// tell waiters to retry locking
	defer r.cond.Broadcast()

	r.cap[item] += count
	return r.cap[item]
}

// Lock locks all items atomically
func (r *Pool) Lock(name ...string) {
	if len(name) == 0 {
		return
	}

	r.cond.L.Lock()
	defer r.cond.L.Unlock()

	for {
		var locked []string
		for _, n := range name {
			if r.lock(n) {
				locked = append(locked, n)
				continue
			}
			// if lock failed, unlock all locked resources
			for _, l := range locked {
				r.unlock(l)
			}
			locked = nil
			break
		}

		if len(locked) == len(name) {
			// lock succeeded!
			return
		}

		// wait for resource update before retrying
		r.cond.Wait()
	}
}

// Unlock unlocks items atomically
func (r *Pool) Unlock(name ...string) {
	if len(name) == 0 {
		return
	}

	r.cond.L.Lock()
	defer r.cond.L.Unlock()

	for _, n := range name {
		r.unlock(n)
	}
	r.cond.Broadcast()
}

func (r *Pool) lock(name string) bool {
	if r.cap[name]-r.locked[name] <= 0 {
		return false
	}

	r.locked[name]++
	return true
}

func (r *Pool) unlock(name string) bool {
	l, ok := r.locked[name]
	switch {
	case l <= 0:
		return false
	case !ok:
		return false
	}
	r.locked[name] = l - 1
	return true
}
