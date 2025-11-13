package scheduler

import (
	"github.com/cryptopunkscc/astrald/astral"
)

// ScheduledTask is an interface for tasks that have been successfully scheduled by the scheduler
type ScheduledTask interface {
	Done
	Task() Task
	State() State
	ScheduledAt() astral.Time
	CancelWithError(error)
	Cancel()
	Err() error
}
