package scheduler

// State represents the state of a scheduled task
type State int64

const (
	StateScheduled State = iota
	StateRunning
	StateDone
)

func (s State) String() string {
	switch s {
	case StateScheduled:
		return "scheduled"
	case StateRunning:
		return "running"
	case StateDone:
		return "done"
	}
	return "invalid"
}
