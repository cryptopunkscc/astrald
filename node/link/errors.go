package link

import "errors"

var ErrPingTimeout = errors.New("ping timeout")
var ErrIdleTimeout = errors.New("idle timeout")
var ErrLinkClosed = errors.New("link closed")
var ErrBufferOverflow = errors.New("buffer overflow")
var ErrRejected = errors.New("query rejected")
var ErrConnClosed = errors.New("connection closed")
