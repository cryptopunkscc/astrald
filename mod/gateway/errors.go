package gateway

import "errors"

var ErrGatewayDenied = errors.New("gateway denied")
var ErrTargetNotReachable = errors.New("target not reachable")
var ErrInvalidGateway = errors.New("invalid gateway")
var ErrSocketUnreachable = errors.New("socket unreachable")
var ErrNodeNotRegistered = errors.New("node not registered")
