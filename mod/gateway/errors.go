package gateway

import "errors"

var ErrUnauthorized = errors.New("unauthorized")
var ErrTargetNotReachable = errors.New("target not reachable")
