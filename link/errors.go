package link

import "errors"

// ErrRejected - the other party rejected the request
var ErrRejected = errors.New("rejected")

// ErrAlreadyClosed - the connection or link is already closed
var ErrAlreadyClosed = errors.New("already closed")
