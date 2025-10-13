package nat

import (
	"context"

	"github.com/cryptopunkscc/astrald/mod/ip"
)

// Puncher is a minimal abstraction for UDP NAT hole punching.
// Implementations supply configuration (e.g., ports, timeouts) outside this interface.
type Puncher interface {
	// Open binds a local UDP socket and returns the chosen local port.
	// The socket is kept by the puncher and reused by HolePunch.
	Open(ctx context.Context) (localPort int, err error)
	// HolePunch attempts a UDP punch towards the given peer IP and port.
	// Returns a PunchResult with an open UDP socket on success.
	HolePunch(ctx context.Context, localIP ip.IP, peer ip.IP, peerPort int) (*PunchResult,
		error)
}

// PunchResult contains the outcome of a successful punch.
type PunchResult struct {
	// NOTE: for now the only supported protocol on top of UDP is UTP (
	// in future we probably will introduce mod/udp package)
	LocalIP   ip.IP
	LocalPort int

	RemoteIP   ip.IP
	RemotePort int
	//
	//	Conn net.PacketConn // open UDP socket ready for I/O
}
