package link

import "errors"

// ErrRejected - the other party rejected the request
var ErrRejected = errors.New("rejected")

// ErrClosed - the connection or link is closed
var ErrClosed = errors.New("connection closed")
