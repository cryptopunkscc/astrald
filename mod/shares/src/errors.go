package shares

import "errors"

var ErrResyncRequired = errors.New("resync required")
var ErrUnavailable = errors.New("unavailable")
var ErrProtocolError = errors.New("protocol error")
