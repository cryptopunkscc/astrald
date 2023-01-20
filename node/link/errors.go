package link

import "errors"

var ErrPingTimeout = errors.New("ping timeout")
var ErrIdleTimeout = errors.New("idle timeout")
