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
	burstInterval   = 25 * time.Millisecond // time between bursts over the whole range
	portGuessRange  = 10                    // number of additional ports to probe around the base port (total ports = 2*portGuessRange + 1)
	packetsPerBurst = 5                     // number of packets sent to each address per burst
	jitterMax       = 5                     // max jitter in milliseconds between rounds
)

// conePuncher is a minimal cone NAT puncher using fixed defaults and a provided peer listen port.
// its simplest form of punching as it only punches port and few ports around
// it. Which wont work on most of home NAT's because they are asymmetric.
// TODO: research statistics about how many NAT's are cone (with and without port preservation) and how many of them are assymmetric.
type conePuncher struct {

	// NOTE: i dont know if this mutex is really needed
	mu        sync.Mutex
	session   []byte         // required session identifier (copied)
	conn      net.PacketConn // bound UDP socket
	localPort int            // cached local port of conn

	// callbacks
	onProbe         func(peer ip.IP, peerPort int, ports []int)
	onSend          func(to *net.UDPAddr, burst int, packets int)
	onProbeReceived func(from *net.UDPAddr)
}

// ConePuncherCallbacks ...
type ConePuncherCallbacks struct {
	OnProbe         func(peer ip.IP, peerPort int, ports []int)
	OnSend          func(to *net.UDPAddr, burst int, packet int)
	OnProbeReceived func(from *net.UDPAddr)
}

// newConePuncherWithSession creates a cone NAT puncher that adopts the provided session.
// Session must be exactly 16 bytes.
func newConePuncherWithSession(session []byte, cb *ConePuncherCallbacks) (puncher nat.Puncher, err error) {
	if len(session) != 16 {
		return nil, fmt.Errorf("session must be 16 bytes")
	}
	s := make([]byte, 16)
	copy(s, session)
	p := &conePuncher{session: s}
	if cb != nil {
		p.onProbe = cb.OnProbe
		p.onSend = cb.OnSend
		p.onProbeReceived = cb.OnProbeReceived
	}
	return p, nil
}

func newConePuncher(cb *ConePuncherCallbacks) (puncher nat.Puncher, err error) {
	session := make([]byte, 16)
	_, err = crand.Read(session)
	if err != nil {
		return nil, fmt.Errorf("generate session: %w", err)
	}
	p := &conePuncher{session: session}
	if cb != nil {
		p.onProbe = cb.OnProbe
		p.onSend = cb.OnSend
		p.onProbeReceived = cb.OnProbeReceived
	}
	return p, nil
}

// Open binds a UDP socket and stores it for later HolePunch use.
func (p *conePuncher) Open() (int, error) {
	if len(p.session) == 0 {
		return 0, errors.New("session is required")
	}

	p.mu.Lock()
	if p.conn != nil {
		lp := p.localPort
		p.mu.Unlock()
		return lp, nil
	}
	p.mu.Unlock()

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
	ports, remoteAddrs, allowed, err := preparePunchTargets(peer, peerPort)
	if err != nil {
		return nil, err
	}
	// callback: plan
	if p.onProbe != nil {
		p.onProbe(peer, peerPort, append([]int(nil), ports...))
	}

	// Create a cancellable context to stop send/receive on success or timeout.
	ctx, cancel := context.WithTimeout(ctx, punchTimeout)
	defer cancel()

	// Channel delivers the first observed remote UDP address (reveals external port).
	recvAddrCh := make(chan *net.UDPAddr, 1)

	// Launch sender and receiver using the snapshot conn.
	go p.receive(ctx, conn, burstInterval, allowed, recvAddrCh)
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
			if p.onProbeReceived != nil {
				p.onProbeReceived(ua)
			}
			select {
			case got <- ua:
			default:
			}
			return
		}
	}
}

// sendBursts sends the session payload to all candidate remoteAddrs simultaneously in each burst, and repeats until ctx is done.
func (p *conePuncher) sendBursts(
	ctx context.Context,
	conn net.PacketConn,
	remoteAddrs []*net.UDPAddr,
	burstEvery time.Duration,
) {
	ticker := time.NewTicker(burstEvery)
	defer ticker.Stop()

	burstIndex := 0
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		for _, remoteAddr := range remoteAddrs {
			if p.onSend != nil {
				p.onSend(remoteAddr, burstIndex, packetsPerBurst)
			}

			for i := 0; i < packetsPerBurst; i++ {
				_, _ = conn.WriteTo(p.session, remoteAddr)
			}
		}

		burstIndex++
		if jitterMax > 0 {
			time.Sleep(jitter(burstEvery, burstEvery*time.Duration(jitterMax)))
		}

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}

	}
}

func candidatePorts(center, spread int) (ports []int) {
	spread = max(spread, 0)
	for i := max(center-spread, 1); i < min(center+spread+1, 65536); i++ {
		ports = append(ports, i)
	}
	return
}

// preparePunchTargets computes candidate ports, resolves remote addresses, and builds the allowed set.
// Returns ports, remoteAddrs, allowed set, and error.
func preparePunchTargets(peer ip.IP, peerPort int) ([]int, []*net.UDPAddr, *sig.Set[string], error) {
	ports := candidatePorts(peerPort, portGuessRange)
	var remoteAddrs []*net.UDPAddr
	for _, rp := range ports {
		addr := net.JoinHostPort(peer.String(), strconv.Itoa(rp))
		remoteAddr, err := net.ResolveUDPAddr("udp", addr)
		if err == nil && remoteAddr != nil {
			remoteAddrs = append(remoteAddrs, remoteAddr)
		}
	}
	if len(remoteAddrs) == 0 {
		return nil, nil, nil, errors.New("no remote addresses to probe")
	}

	// Build set of allowed remote endpoints for validation.
	var allowed sig.Set[string]
	for _, remoteAddr := range remoteAddrs {
		allowed.Add(remoteAddr.String())
	}

	return ports, remoteAddrs, &allowed, nil
}

// Jitter adds uniform jitter: [-spread .. +spread]
func jitter(base, spread time.Duration) time.Duration {
	if spread <= 0 {
		return base
	}
	delta := rand.Int63n(int64(spread)*2+1) - int64(spread)
	out := int64(base) + delta
	if out < 0 {
		return 0
	}
	return time.Duration(out)
}
