package activity

import (
	"context"
	"time"
)

type Tracker interface {
	Touch()
	AddBytesRead(int)
	AddBytesWritten(int)
}

type Activity struct {
	parent       Tracker
	bytesRead    int
	bytesWritten int
	lastActivity time.Time
	sticky       bool
}

func New(parent Tracker) *Activity {
	return &Activity{
		parent:       parent,
		lastActivity: time.Now(),
	}
}

func (a *Activity) Touch() {
	a.lastActivity = time.Now()
	if a.parent != nil {
		a.parent.Touch()
	}
}

func (a *Activity) Idle() time.Duration {
	if a.sticky {
		return 0
	}
	return time.Now().Sub(a.lastActivity)
}

func (a *Activity) AddBytesRead(n int) {
	a.bytesRead += n
	if a.parent != nil {
		a.parent.AddBytesRead(n)
	}
}

func (a *Activity) AddBytesWritten(n int) {
	a.bytesWritten += n
	if a.parent != nil {
		a.parent.AddBytesWritten(n)
	}
}

func (a *Activity) BytesRead() int {
	return a.bytesRead
}

func (a *Activity) BytesWritten() int {
	return a.bytesWritten
}

func (a *Activity) SetParent(parent Tracker) {
	a.parent = parent
}

func (a *Activity) WaitIdle(ctx context.Context, timeout time.Duration) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(timeout - a.Idle()):
			if a.Idle() >= timeout {
				return nil
			}
		}
	}
}

func (a *Activity) SetSticky(sticky bool) {
	a.sticky = sticky
	if sticky == false {
		a.Touch()
	}
}
