package sync

import "time"

type Waiter interface {
	Wait() <-chan struct{}
}

type Idler interface {
	Idle() time.Duration
}

type Notifier interface {
	Notify()
}

type Flagger interface {
	State() bool
}
