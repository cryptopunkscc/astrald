package gateway

import "errors"

var ErrUnauthorized = errors.New("unauthorized")
var ErrTargetNotReachable = errors.New("target not reachable")
var ErrInvalidGateway = errors.New("invalid gateway")
var ErrSocketUnreachable = errors.New("socket unreachable")
