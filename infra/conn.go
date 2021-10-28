package infra

import "io"

type Conn interface {
	io.ReadWriteCloser // Basic IO operations
	Outbound() bool    // Returns true if we are the active party, false otherwise
	LocalAddr() Addr   // Returns local network address if known, nil otherwise
	RemoteAddr() Addr  // Returns the other party's network address if known, nil otherwise
}
