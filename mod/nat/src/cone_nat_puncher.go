package nat

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/ip"
	"github.com/cryptopunkscc/astrald/mod/nat"
)

// Ensure conePuncher implements the public Puncher interface from the root package.
var _ nat.Puncher = (*conePuncher)(nil)

const (
	punchTimeout  = 10 * time.Second // ~1.5s total per attempt
	burstInterval = 50 * time.
			Millisecond // time between bursts over the whole range
	packetSpacing   = 25 * time.Millisecond // 20–80ms between packets
	portGuessRange  = 10                    // try [-5..+5] around base port (including itself)
	packetsPerBurst = 5                     // number of packets sent to each address per burst
)

// conePuncher is a minimal cone NAT puncher using fixed defaults and a provided peer listen port.
// its simplest form of punching as it only punches port and few ports around
// it. Which wont work on most of home NAT's because they are asymmetric.
type conePuncher struct {
	mu        sync.Mutex
	session   []byte         // required session identifier (copied)
	conn      net.PacketConn // bound UDP socket
	localPort int            // cached local port of conn
	opened    bool           // whether conn is opened
}

// newConePuncher creates a cone NAT puncher that targets the given peer listen port.
// Session is required and must be non-empty.
func newConePuncher() (puncher nat.Puncher, err error) {
	session := make([]byte, 16)

	_, err = rand.Read(session)
	if err != nil {
		return nil, fmt.Errorf("nat: generate session: %w", err)
	}

	return &conePuncher{session: session}, nil
}

// Open binds a UDP socket and stores it for later HolePunch use.
func (p *conePuncher) Open(ctx context.Context) (int, error) {
	if len(p.session) == 0 {
		return 0, errors.New("nat: session is required")
	}
	// Fast path: check under lock
	p.mu.Lock()
	if p.opened && p.conn != nil {
		lp := p.localPort
		p.mu.Unlock()
		return lp, nil
	}
	p.mu.Unlock()

	// Not opened yet — create socket without holding lock to avoid blocking others.
	conn, err := net.ListenPacket("udp", ":0")
	if err != nil {
		return 0, fmt.Errorf("nat: listen udp: %w", err)
	}
	lp := 0
	if ua, ok := conn.LocalAddr().(*net.UDPAddr); ok {
		lp = ua.Port
	}

	p.mu.Lock()
	defer p.mu.Unlock()
	if p.opened && p.conn != nil {
		// another goroutine opened; use existing and close ours
		_ = conn.Close()
		return p.localPort, nil
	}
	p.conn = conn
	p.localPort = lp
	p.opened = true
	return lp, nil
}

// Close releases any resources held by the puncher (open sockets).
func (p *conePuncher) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.conn != nil {
		err := p.conn.Close()
		p.conn = nil
		p.opened = false
		p.localPort = 0
		return err
	}
	p.opened = false
	p.localPort = 0
	return nil
}

func (p *conePuncher) HolePunch(
	ctx context.Context,
	peer ip.IP,
	peerPort int,
) (*nat.PunchResult, error) {
	if peer == nil || peer.String() == "" {
		return nil, errors.New("nat: empty peer IP")
	}
	if len(p.session) == 0 {
		return nil, errors.New("nat: session is required")
	}
	// Ensure socket is opened (Open is thread-safe)
	if _, err := p.Open(ctx); err != nil {
		return nil, err
	}

	// Snapshot conn and localPort under lock
	p.mu.Lock()
	conn := p.conn
	localPort := p.localPort
	p.mu.Unlock()

	if conn == nil {
		return nil, errors.New("nat: no UDP connection available")
	}

	// Prepare candidate remote addresses around the peer's reported port.
	var raddrs []net.Addr
	for _, rp := range candidatePorts(peerPort, portGuessRange) {
		addr := net.JoinHostPort(peer.String(), strconv.Itoa(rp))
		ra, err := net.ResolveUDPAddr("udp", addr)
		if err == nil && ra != nil {
			raddrs = append(raddrs, ra)
		}
	}
	if len(raddrs) == 0 {
		return nil, errors.New("nat: no remote addresses to probe")
	}

	// Create a cancellable context to stop send/receive on success or timeout.
	ctx, cancel := context.WithTimeout(ctx, punchTimeout)
	defer cancel()

	// Channel delivers the first observed remote UDP address (reveals external port).
	recvAddrCh := make(chan *net.UDPAddr, 1)

	// Launch sender and receiver using the snapshot conn.
	go p.receive(ctx, conn, burstInterval, recvAddrCh)
	go p.sendBursts(ctx, conn, raddrs, burstInterval, packetSpacing)

	// Wait for success or timeout/cancel.
	select {
	case ra := <-recvAddrCh:
		return &nat.PunchResult{
			LocalPort:  astral.Uint16(localPort),
			RemoteIP:   peer,
			RemotePort: astral.Uint16(ra.Port),
		}, nil
	case <-ctx.Done():
		// Close and reset on failure to avoid leaking the socket
		p.mu.Lock()
		if p.conn != nil {
			_ = p.conn.Close()
			p.conn = nil
		}
		p.opened = false
		p.localPort = 0
		p.mu.Unlock()
		return nil, ctx.Err()
	}
}

func (p *conePuncher) Session() []byte {
	return p.session
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

// sendBursts sends the session payload to all candidate raddrs simultaneously in each burst, and repeats until ctx is done.
func (p *conePuncher) sendBursts(ctx context.Context, conn net.PacketConn, raddrs []net.Addr, burstEvery, spacing time.Duration) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		for i := 0; i < packetsPerBurst; i++ {
			for _, ra := range raddrs {
				_, _ = conn.WriteTo(p.session, ra)
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
