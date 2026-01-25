package astral

import (
	"errors"
)

// simple errors

// ErrTimeout - query timed out
var ErrTimeout = errors.New("query timeout")

// ErrZoneExcluded - operation requires zones excluded from the scope
var ErrZoneExcluded = errors.New("zone excluded")

// ErrTargetNotAllowed - target was blocked by a policy or a filter
var ErrTargetNotAllowed = errors.New("target not allowed")
