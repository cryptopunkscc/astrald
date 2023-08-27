package link

import "errors"

var ErrProtocolError = errors.New("protocol error")
var ErrLinkClosed = errors.New("link closed")
var ErrLinkClosedByPeer = errors.New("link closed by peer")
var ErrRemoteBufferOverflow = errors.New("remote buffer overflow")
var ErrPortBufferOverflow = errors.New("port buffer overflow")
var ErrPortBufferEmpty = errors.New("port buffer empty")
var ErrPortClosed = errors.New("port closed")
var ErrPingTimeout = errors.New("ping timeout")
var ErrTooManyPings = errors.New("too many pings in progress")
var ErrInvalidNonce = errors.New("invalid ping nonce")
