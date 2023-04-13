package link

import "errors"

var ErrPingTimeout = errors.New("ping timeout")
var ErrIdleTimeout = errors.New("idle timeout")
var ErrLinkClosed = errors.New("link closed")
var ErrQueryTimeout = errors.New("query timedout")
var ErrBufferOverflow = errors.New("buffer overflow")
var ErrQueryFinished = errors.New("query finished")
