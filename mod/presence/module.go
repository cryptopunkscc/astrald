package presence

import (
	"bytes"
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/infra/inet"
	"github.com/cryptopunkscc/astrald/infra/ip"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/event"
	"github.com/cryptopunkscc/astrald/sig"
	"net"
	"strconv"
	"sync"
	"time"
)

type Module struct {
	node   node.Node
	config Config

	presenceConn *net.UDPConn

	entries map[string]*entry
	mu      sync.Mutex
	events  event.Queue
	skip    map[string]struct{}
}

const defaultPresencePort = 8829
const announceInterval = 1 * time.Minute

type presence struct {
	Identity id.Identity
	Port     int
	Flags    uint8
}

const (
	flagNone     = 0x00
	flagDiscover = 0x01
	flagBye      = 0x02
)

const presenceCSLQ = "x61 x70 x00 x00 v s c"

func (p *presence) MarshalCSLQ(enc *cslq.Encoder) error {
	return enc.Encode(presenceCSLQ, p.Identity, p.Port, p.Flags)
}

func (p *presence) UnmarshalCSLQ(dec *cslq.Decoder) error {
	return dec.Decode(presenceCSLQ, &p.Identity, &p.Port, &p.Flags)
}

func (m *Module) Run(ctx context.Context) (err error) {
	ctx, shutdown := context.WithCancel(ctx)

	var errCh = make(chan error, 2)
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()

		discover, err := m.Discover(ctx)
		if err != nil {
			errCh <- err
			return
		}

		for presence := range discover {
			hex := presence.Identity.PublicKeyHex()
			if _, skip := m.skip[hex]; skip {
				continue
			}

			m.handle(ctx, presence)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		err := m.Announce(ctx)
		if err != nil {
			errCh <- err
		}
	}()

	go func() {
		select {
		case <-ctx.Done():
		case err = <-errCh:
			shutdown()
		}
	}()

	wg.Wait()

	return nil
}

func (m *Module) Identities() <-chan id.Identity {
	m.mu.Lock()
	defer m.mu.Unlock()

	ch := make(chan id.Identity, len(m.entries))
	for hex := range m.entries {
		i, err := id.ParsePublicKeyHex(hex)
		if err != nil {
			panic(err)
		}
		ch <- i
	}
	close(ch)

	return ch
}

func (m *Module) ignore(identity id.Identity) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.skip[identity.PublicKeyHex()] = struct{}{}
}

func (m *Module) unignore(identity id.Identity) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.skip, identity.PublicKeyHex())
}

func (m *Module) handle(ctx context.Context, ip Presence) {
	m.mu.Lock()
	defer m.mu.Unlock()

	hex := ip.Identity.PublicKeyHex()

	if e, found := m.entries[hex]; found {
		e.Touch()
		return
	}

	e := trackPresence(ctx, ip)
	m.entries[hex] = e

	// remove presence entry when it times out
	sig.OnCtx(ctx, e, func() {
		m.remove(hex)
	})

	log.Tag("presence").Info("%s present on %s", ip.Identity, log.Em(ip.Addr.Network()))

	m.events.Emit(EventIdentityPresent{ip.Identity, ip.Addr})

	_ = m.node.Tracker().Add(ip.Identity, ip.Addr, time.Now().Add(60*time.Minute))
}

func (m *Module) Discover(ctx context.Context) (<-chan Presence, error) {
	// check presence socket
	if err := m.setupPresenceConn(); err != nil {
		return nil, err
	}

	go func() {
		<-ctx.Done()
		_ = m.presenceConn.Close()
	}()

	outCh := make(chan Presence)

	go func() {
		defer close(outCh)
		buf := make([]byte, 1024)

		for {
			n, srcAddr, err := m.presenceConn.ReadFromUDP(buf)
			if err != nil {
				return
			}

			var p presence
			var r = bytes.NewReader(buf[:n])

			if err := cslq.Decode(r, "v", &p); err != nil {
				continue
			}

			if p.Identity.IsEqual(m.node.Identity()) {
				// ignore our own presence
				continue
			}

			addr, err := inet.Parse(srcAddr.IP.String() + ":" + strconv.Itoa(p.Port))
			if err != nil {
				panic(err)
			}

			outCh <- Presence{
				Identity: p.Identity,
				Addr:     addr,
				Present:  p.Flags&flagBye == 0,
			}

			if p.Flags&flagDiscover != 0 {
				m.sendPresence(srcAddr, presence{
					Identity: m.node.Identity(),
					Port:     m.getListenPort(),
					Flags:    flagNone,
				})
			}
		}
	}()

	return outCh, nil
}

func (m *Module) remove(hex string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if e, found := m.entries[hex]; found {
		delete(m.entries, hex)

		log.Tag("presence").Info("%s gone from %s", e.id, log.Em(e.addr.Network()))

		m.events.Emit(EventIdentityGone{e.id})
	}
}

func (inet *Module) Announce(ctx context.Context) error {
	if err := inet.broadcastPresence(&presence{
		Identity: inet.node.Identity(),
		Port:     inet.getListenPort(),
		Flags:    flagDiscover,
	}); err != nil {
		return err
	}

	log.Log("announcing presence")

	go func() {
		for {
			select {
			case <-time.After(announceInterval):
				if err := inet.broadcastPresence(&presence{
					Identity: inet.node.Identity(),
					Port:     inet.getListenPort(),
					Flags:    flagNone,
				}); err != nil {
					log.Error("announce: %s", err)
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return nil
}

func (inet *Module) broadcastPresence(p *presence) error {
	// check presence socket
	if err := inet.setupPresenceConn(); err != nil {
		return err
	}

	// prepare packet data
	var packet = &bytes.Buffer{}
	if err := cslq.Encode(packet, "v", p); err != nil {
		return err
	}

	ifaces, err := net.Interfaces()
	if err != nil {
		return err
	}

	for _, iface := range ifaces {
		if isInterfaceEnabled(iface) {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			return err
		}

		for _, addr := range addrs {
			broadcastIP, err := ip.BroadcastAddr(addr)
			if err != nil {
				return err
			}

			if ip.IsLinkLocal(broadcastIP) {
				continue
			}

			var broadcastAddr = net.UDPAddr{
				IP:   broadcastIP,
				Port: defaultPresencePort,
			}

			_, err = inet.presenceConn.WriteTo(packet.Bytes(), &broadcastAddr)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func isInterfaceEnabled(iface net.Interface) bool {
	return (iface.Flags&net.FlagUp != 0) &&
		(iface.Flags&net.FlagBroadcast != 0) &&
		(iface.Flags&net.FlagLoopback == 0)
}

func (inet *Module) getListenPort() int {
	return inet.node.Infra().Inet().ListenPort()
}

func (inet *Module) setupPresenceConn() (err error) {
	inet.mu.Lock()
	defer inet.mu.Unlock()

	// already set up?
	if inet.presenceConn != nil {
		return nil
	}

	// resolve local address
	var localAddr *net.UDPAddr
	localAddr, err = net.ResolveUDPAddr("udp", ":"+strconv.Itoa(defaultPresencePort))
	if err != nil {
		return
	}

	// bind the udp socket
	inet.presenceConn, err = net.ListenUDP("udp", localAddr)
	return
}

func (inet *Module) sendPresence(destAddr *net.UDPAddr, p presence) (err error) {
	// check presence socket
	if err = inet.setupPresenceConn(); err != nil {
		return
	}

	// prepare packet data
	var packet = &bytes.Buffer{}
	if err = cslq.Encode(packet, "v", p); err != nil {
		return
	}

	// send message
	_, err = inet.presenceConn.WriteTo(packet.Bytes(), destAddr)
	return
}
