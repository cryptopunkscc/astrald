package link

import "errors"

// ErrRejected - the other party rejected the request
var ErrRejected = errors.New("rejected")

// ErrStreamClosed - link is closed and cannot be written to
var ErrStreamClosed = errors.New("link closed")
