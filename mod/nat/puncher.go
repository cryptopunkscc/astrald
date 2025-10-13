package nat

import (
	"context"
	"net"

	"github.com/cryptopunkscc/astrald/mod/ip"
	"github.com/cryptopunkscc/astrald/mod/utp"
)

// Puncher is a minimal abstraction for UDP NAT hole punching.
// Implementations supply configuration (e.g., ports, timeouts) outside this interface.
type Puncher interface {
	// HolePunch attempts a UDP punch towards the given peer IP and port.
	// Returns a PunchResult with an open UDP socket on success.
	HolePunch(ctx context.Context, peer ip.IP, peerPort int) (*PunchResult, error)
}

// PunchResult contains the outcome of a successful punch.
type PunchResult struct {
	// NOTE: for now the only supported protocol on top of UDP is UTP (
	// in future we probably will introduce mod/udp package)
	Local  utp.Endpoint // resolved local endpoint (bound UDP port)
	Remote utp.Endpoint // target peer endpoint (IP + port used by the puncher)
	//
	Conn net.PacketConn // open UDP socket ready for I/O
}
