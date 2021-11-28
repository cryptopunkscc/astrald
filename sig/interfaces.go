package sig

import "time"

type Waiter interface {
	Wait() <-chan struct{}
}

type Idler interface {
	Idle() time.Duration
}
