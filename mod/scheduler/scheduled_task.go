package scheduler

import (
	"time"
)

// ScheduledTask is an interface for tasks that have been successfully scheduled by the scheduler
type ScheduledTask interface {
	Done
	Task() Task
	State() State
	ScheduledAt() time.Time
	// CancelWithError cancels the task and sets the error returned by Err.
	CancelWithError(error)
	Cancel()
	Err() error
}
