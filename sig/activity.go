package sig

import (
	"sync"
	"time"
)

// Activity tracks idle time since last activity.
type Activity struct {
	lastActivity time.Time
	ongoing      int
	mu           sync.Mutex
}

// Idle returns the duration since last activity. For ongoing activities it's always 0. If there was never
// any activity, it returns -1.
func (a *Activity) Idle() time.Duration {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.lastActivity.IsZero() {
		return -1
	}

	if a.ongoing > 0 {
		return 0
	}
	return time.Now().Sub(a.lastActivity)
}

// Add increments the ongoing counter by the argument. An activity with a positive ongoing counter is never idle.
func (a *Activity) Add(ongoing int) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.ongoing += ongoing
	a.touch()
}

// Done decrements the ongoing counter by 1.
func (a *Activity) Done() {
	a.Add(-1)
}

// Touch resets idle time
func (a *Activity) Touch() {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.touch()
}

func (a *Activity) touch() {
	a.lastActivity = time.Now()
}
