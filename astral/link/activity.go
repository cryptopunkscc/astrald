package link

import (
	"context"
	"time"
)

type ActivityTracker interface {
	Touch()
	AddBytesRead(int)
	AddBytesWritten(int)
}

type Activity struct {
	activityHost ActivityTracker
	bytesRead    int
	bytesWritten int
	lastActivity time.Time
}

func NewActivity(activityHost ActivityTracker) *Activity {
	return &Activity{
		activityHost: activityHost,
		lastActivity: time.Now(),
	}
}

func (a *Activity) Touch() {
	a.lastActivity = time.Now()
	if a.activityHost != nil {
		a.activityHost.Touch()
	}
}

func (a *Activity) Idle() time.Duration {
	return time.Now().Sub(a.lastActivity)
}

func (a *Activity) AddBytesRead(n int) {
	a.bytesRead += n
	if a.activityHost != nil {
		a.activityHost.AddBytesRead(n)
	}
}

func (a *Activity) AddBytesWritten(n int) {
	a.bytesWritten += n
	if a.activityHost != nil {
		a.activityHost.AddBytesWritten(n)
	}
}

func (a *Activity) BytesRead() int {
	return a.bytesRead
}

func (a *Activity) BytesWritten() int {
	return a.bytesWritten
}

func (a *Activity) SetActivityHost(activityHost ActivityTracker) {
	a.activityHost = activityHost
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
