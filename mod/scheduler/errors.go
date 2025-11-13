package scheduler

import "errors"

var (
	ErrNotRunning = errors.New("scheduler not running")
	ErrTaskIsNil  = errors.New("task is nil")
)
