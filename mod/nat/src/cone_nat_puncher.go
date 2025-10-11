package nat

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/ip"
	"github.com/cryptopunkscc/astrald/mod/nat"
	"github.com/cryptopunkscc/astrald/mod/utp"
)

// NOTE: Mostly AI will be polished

// Ensure conePuncher implements the public Puncher interface from the root package.
var _ nat.Puncher = (*conePuncher)(nil)

const (
	punchTimeout   = 1500 * time.Millisecond // ~1.5s total per attempt
	burstInterval  = 100 * time.Millisecond  // time between bursts over the whole range
	packetSpacing  = 30 * time.Millisecond   // 20–80ms between packets
	portGuessRange = 10                      // try [-5..+5] around base port (including itself)
)

// conePuncher is a minimal cone NAT puncher using fixed defaults and a provided peer listen port.
type conePuncher struct {
	listenPort int    // peer private listen port hint; if 0, fallback to local port as center
	session    []byte // required session identifier
}

// newConePuncher creates a cone NAT puncher that targets the given peer listen port.
// Session is required and must be non-empty.
func newConePuncher(listenPort int, session []byte) nat.Puncher {
	if len(session) == 0 {
		panic("nat: session is required for cone puncher")
	}
	return &conePuncher{listenPort: listenPort, session: append([]byte(nil), session...)}
}

func (p *conePuncher) HolePunch(ctx context.Context, peer ip.IP) (*nat.PunchResult, error) {
	if peer == nil || peer.String() == "" {
		return nil, errors.New("nat: empty peer IP")
	}
	if p.listenPort < 0 || p.listenPort > 65535 {
		return nil, errors.New("nat: invalid listenPort")
	}
	if len(p.session) == 0 {
		return nil, errors.New("nat: session is required")
	}

	// Bind local UDP socket on random port (Option B allows arbitrary L; could be configured later).
	conn, err := net.ListenPacket("udp", ":0")
	if err != nil {
		return nil, fmt.Errorf("nat: listen udp: %w", err)
	}
	success := false
	defer func() {
		if !success {
			_ = conn.Close()
		}
	}()

	// Determine local port and select base remote port center.
	localPort := 0
	if ua, ok := conn.LocalAddr().(*net.UDPAddr); ok {
		localPort = ua.Port
	}
	base := p.listenPort
	if base == 0 {
		base = localPort // fallback: use our local port as the probing center
	}

	// Prepare candidate remote addresses around the base port.
	var raddrs []net.Addr
	for _, rp := range candidatePorts(base, portGuessRange) {
		ra, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", peer.String(), rp))
		if err == nil && ra != nil {
			raddrs = append(raddrs, ra)
		}
	}
	if len(raddrs) == 0 {
		return nil, errors.New("nat: no remote addresses to probe")
	}

	localEP := utp.Endpoint{IP: ip.IP(nil), Port: astral.Uint16(localPort)}

	// Create a cancellable context to stop send/receive on success or timeout.
	ctx, cancel := context.WithTimeout(ctx, punchTimeout)
	defer cancel()

	// Channel delivers the first observed remote UDP address (reveals external port).
	recvAddrCh := make(chan *net.UDPAddr, 1)

	// Launch sender and receiver.
	go p.receive(ctx, conn, burstInterval, recvAddrCh)
	go p.sendBursts(ctx, conn, raddrs, burstInterval, packetSpacing)

	// Wait for success or timeout/cancel.
	select {
	case ra := <-recvAddrCh:
		success = true
		// stop background goroutines
		cancel()
		remoteEP := utp.Endpoint{IP: peer, Port: astral.Uint16(ra.Port)}
		return &nat.PunchResult{Local: localEP, Remote: remoteEP, Conn: conn}, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// receive listens for any incoming UDP probe and reports the sender address if payload matches our session.
func (p *conePuncher) receive(ctx context.Context, conn net.PacketConn, interval time.Duration, got chan<- *net.UDPAddr) {
	buf := make([]byte, 1500)
	for {
		_ = conn.SetReadDeadline(time.Now().Add(interval))
		n, addr, err := conn.ReadFrom(buf)
		if err == nil && n == len(p.session) && bytes.Equal(buf[:n], p.session) {
			if ua, ok2 := addr.(*net.UDPAddr); ok2 {
				select {
				case got <- ua:
				default:
				}
				return
			}
		}
		select {
		case <-ctx.Done():
			return
		default:
		}
	}
}

// sendBursts sends the session payload to all candidate raddrs, spacing packets slightly, and repeats in bursts until ctx is done.
func (p *conePuncher) sendBursts(ctx context.Context, conn net.PacketConn, raddrs []net.Addr, burstEvery, spacing time.Duration) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		for _, ra := range raddrs {
			_, _ = conn.WriteTo(p.session, ra)
			t := time.NewTimer(spacing)
			select {
			case <-ctx.Done():
				if !t.Stop() {
					<-t.C
				}
				return
			case <-t.C:
			}
		}
		// wait until next burst
		b := time.NewTimer(burstEvery)
		select {
		case <-ctx.Done():
			if !b.Stop() {
				<-b.C
			}
			return
		case <-b.C:
		}
	}
}

// candidatePorts returns [center, center-1, center+1, center-2, center+2, ...] within [1..65535].
func candidatePorts(center, spread int) []int {
	if spread < 0 {
		spread = 0
	}
	res := make([]int, 0, 1+2*spread)
	add := func(p int) {
		if p >= 1 && p <= 65535 {
			res = append(res, p)
		}
	}
	add(center)
	for d := 1; d <= spread/2; d++ {
		add(center - d)
		add(center + d)
	}
	// If spread is odd, include the last side
	if spread%2 == 1 {
		d := spread/2 + 1
		add(center - d)
		add(center + d)
	}
	return res
}
