package link

import "errors"

// ErrRejected - the other party rejected the request
var ErrRejected = errors.New("rejected")

// ErrClosed - the connection or link is closed
var ErrClosed = errors.New("connection closed")

// ErrIOTimeout - the remote host failed to respond within a time limit
var ErrIOTimeout = errors.New("i/o timeout")
