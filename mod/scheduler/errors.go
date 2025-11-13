package scheduler

import "errors"

var (
	ErrNotRunning   = errors.New("scheduler not running")
	ErrInvalidState = errors.New("invalid state")
	ErrTaskIsNil    = errors.New("task is nil")
	ErrContextIsNil = errors.New("context is nil")
)
