package routing

import (
	"time"

	"github.com/cryptopunkscc/astrald/astral"
)

// Report holds information about a finished op call
type Report struct {
	Query *astral.Query
	Time  time.Duration
	Err   error
}
