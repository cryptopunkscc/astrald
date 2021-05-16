package link

import "errors"

// ErrRejected - the other party rejected the request
var ErrRejected = errors.New("rejected")

// ErrLinkClosed - link is closed and cannot be written to
var ErrLinkClosed = errors.New("link closed")
