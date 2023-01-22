package sig

import "time"

type Signal interface {
	Done() <-chan struct{}
}

type Idler interface {
	Idle() time.Duration
}
