package nat

import (
	"bytes"
	"context"
	crand "crypto/rand"
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
	"github.com/cryptopunkscc/astrald/sig"
)

// Ensure conePuncher implements the public Puncher interface from the root package.
var _ nat.Puncher = (*conePuncher)(nil)

const (
	punchTimeout    = 10 * time.Second      // total timeout per attempt
	burstInterval   = 50 * time.Millisecond // time between bursts over the whole range
	portGuessRange  = 5                     // number of additional ports to probe around the base port (total ports = 2*portGuessRange + 1)
	packetsPerBurst = 5                     // number of packets sent to each address per burst
	jitterMax       = 3                     // max jitter in milliseconds between rounds
)

// conePuncher is a minimal cone NAT puncher using fixed defaults and a provided peer listen port.
// its simplest form of punching as it only punches port and few ports around
// it. Which wont work on most of home NAT's because they are asymmetric.
type conePuncher struct {
	mu        sync.Mutex
	session   []byte         // required session identifier (copied)
	conn      net.PacketConn // bound UDP socket
	localPort int            // cached local port of conn
}

// newConePuncher creates a cone NAT puncher with a new randomly generated session.
func newConePuncher() (puncher nat.Puncher, err error) {
	session := make([]byte, 16)
	_, err = crand.Read(session)
	if err != nil {
		return nil, fmt.Errorf("generate session: %w", err)
	}
	return &conePuncher{session: session}, nil
}

// newConePuncherWithSession creates a cone NAT puncher that adopts the provided session.
// Session must be exactly 16 bytes.
func newConePuncherWithSession(session []byte) (puncher nat.Puncher, err error) {
	if len(session) != 16 {
		return nil, errors.New("session must be 16 bytes")
	}
	// defensive copy of session
	s := make([]byte, 16)
	copy(s, session)
	return &conePuncher{session: s}, nil
}

// Open binds a UDP socket and stores it for later HolePunch use.
func (p *conePuncher) Open() (int, error) {
	if len(p.session) == 0 {
		return 0, errors.New("session is required")
	}
	// Fast path: check under lock
	p.mu.Lock()
	if p.conn != nil {
		lp := p.localPort
		p.mu.Unlock()
		return lp, nil
	}
	p.mu.Unlock()

	// Not opened yet — create socket without holding lock to avoid blocking others.
	conn, err := net.ListenPacket("udp", ":0")
	if err != nil {
		return 0, fmt.Errorf("listen udp: %w", err)
	}
	lp := 0
	if ua, ok := conn.LocalAddr().(*net.UDPAddr); ok {
		lp = ua.Port
	}

	p.mu.Lock()
	defer p.mu.Unlock()
	if p.conn != nil {
		// another goroutine opened; use existing and close ours
		_ = conn.Close()
		return p.localPort, nil
	}
	p.conn = conn
	p.localPort = lp
	return lp, nil
}

// Close releases any resources held by the puncher (open sockets).
func (p *conePuncher) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.conn != nil {
		err := p.conn.Close()
		p.conn = nil
		p.localPort = 0
		return err
	}
	p.localPort = 0
	return nil
}

// HolePunch attempts to punch a hole through NAT by sending UDP packets to candidate ports around peerPort.
// On success, returns the observed remote IP and port from the first matching incoming packet,
// which may differ from the announced peerIP/peerPort due to NAT translation.
func (p *conePuncher) HolePunch(
	ctx context.Context,
	peer ip.IP,
	peerPort int,
) (*nat.PunchResult, error) {
	if peer == nil || peer.String() == "" {
		return nil, errors.New("empty peer IP")
	}
	if len(p.session) == 0 {
		return nil, errors.New("session is required")
	}
	if peerPort < 1 || peerPort > 65535 {
		return nil, fmt.Errorf("invalid peer port: %d", peerPort)
	}

	// Snapshot conn and localPort under lock
	p.mu.Lock()
	conn := p.conn
	localPort := p.localPort
	p.mu.Unlock()

	if conn == nil {
		return nil, errors.New("no UDP connection available")
	}

	// Prepare candidate remote addresses around the peer's reported port.
	var remoteAddrs []*net.UDPAddr
	for _, rp := range candidatePorts(peerPort, portGuessRange) {
		addr := net.JoinHostPort(peer.String(), strconv.Itoa(rp))
		remoteAddr, err := net.ResolveUDPAddr("udp", addr)
		if err == nil && remoteAddr != nil {
			remoteAddrs = append(remoteAddrs, remoteAddr)
		}
	}
	if len(remoteAddrs) == 0 {
		return nil, errors.New("no remote addresses to probe")
	}

	// Build set of allowed remote endpoints for validation.
	var allowed sig.Set[string]
	for _, remoteAddr := range remoteAddrs {
		_ = allowed.Add(remoteAddr.String())
	}

	// Create a cancellable context to stop send/receive on success or timeout.
	ctx, cancel := context.WithTimeout(ctx, punchTimeout)
	defer cancel()

	// Channel delivers the first observed remote UDP address (reveals external port).
	recvAddrCh := make(chan *net.UDPAddr, 1)

	// Launch sender and receiver using the snapshot conn.
	go p.receive(ctx, conn, burstInterval, &allowed, recvAddrCh)
	go p.sendBursts(ctx, conn, remoteAddrs, burstInterval)

	// Wait for success or timeout/cancel.
	select {
	case remoteAddr := <-recvAddrCh:
		return &nat.PunchResult{
			LocalPort:  astral.Uint16(localPort),
			RemoteIP:   ip.IP(remoteAddr.IP),
			RemotePort: astral.Uint16(remoteAddr.Port),
		}, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (p *conePuncher) Session() []byte {
	return append([]byte(nil), p.session...)
}

func (p *conePuncher) LocalPort() int {
	return p.localPort
}

// receive listens for any incoming UDP probe and reports the sender address if payload matches our session.
func (p *conePuncher) receive(ctx context.Context, conn net.PacketConn, interval time.Duration, allowed *sig.Set[string], got chan<- *net.UDPAddr) {
	buf := make([]byte, 1500)
	for {
		_ = conn.SetReadDeadline(time.Now().Add(interval))
		n, addr, err := conn.ReadFrom(buf)
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return
			}
			if ne, ok := err.(net.Error); ok && ne.Timeout() {
				continue
			}
			select {
			case <-ctx.Done():
				return
			default:
				time.Sleep(5 * time.Millisecond)
				continue
			}
		}
		ua, ok := addr.(*net.UDPAddr)
		if !ok {
			continue
		}
		if !allowed.Contains(ua.String()) {
			continue
		}
		if n == len(p.session) && bytes.Equal(buf[:n], p.session) {
			select {
			case got <- ua:
			default:
			}
			return
		}
	}
}

// sendBursts sends the session payload to all candidate remoteAddrs simultaneously in each burst, and repeats until ctx is done.
func (p *conePuncher) sendBursts(ctx context.Context, conn net.PacketConn, remoteAddrs []*net.UDPAddr, burstEvery time.Duration) {
	ticker := time.NewTicker(burstEvery)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		for i := 0; i < packetsPerBurst; i++ {
			for _, remoteAddr := range remoteAddrs {
				_, _ = conn.WriteTo(p.session, remoteAddr)
			}
			// add micro-jitter between rounds
			jitter := time.Duration(rand.Intn(jitterMax*2+1)-jitterMax) * time.Millisecond
			time.Sleep(jitter)
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
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
