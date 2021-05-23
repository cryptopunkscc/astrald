package net

import "io"

// Conn represents a generic interface for any type of bidirectional stream of bytes that preserves order
// and gaurantees data integrity. No authentication or encryption is guaranteed. Identity of the other party is
// unknown.
type Conn interface {
	io.ReadWriteCloser // Basic IO operations
	Outbound() bool    // Returns true if we are the active party, false otherwise
	RemoteAddr() Addr  // Returns the other party's network address if knwon, nil otherwise
}
