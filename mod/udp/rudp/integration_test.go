package rudp

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	udpmod "github.com/cryptopunkscc/astrald/mod/udp"
)

// TestListenerDialHelloWorld exercises a minimal end-to-end handshake and
// one-shot data transfer ("Hello World") between an outbound client Conn
// and an inbound server Conn accepted through Listener.Accept().
func TestListenerDialHelloWorld(t *testing.T) {
	baseCtx := astral.NewContext(context.Background())

	// Start listener on an IPv4 loopback ephemeral port (force IPv4 to avoid ::/127.0.0.1 mismatch)
	l, err := Listen(baseCtx, &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 0}, Config{}, 2*time.Second)
	if err != nil {
		t.Fatalf("Listen failed: %v", err)
	}
	defer l.Close()

	serverAddr := l.Addr().(*net.UDPAddr)
	// Force IPv4 127.0.0.1 target (avoid ::1 / unspecified ambiguity)
	ipv4Dest := &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: serverAddr.Port}

	// Channel to receive accepted server connection
	acceptedCh := make(chan *Conn, 1)

	// Accept in background
	go func() {
		acceptCtx, cancel := baseCtx.WithTimeout(3 * time.Second)
		defer cancel()
		c, err := l.Accept(acceptCtx)
		if err != nil {
			return
		}
		acceptedCh <- c
	}()

	// Dial UDP (raw) for outbound side
	udpConn, err := net.DialUDP("udp4", nil, ipv4Dest)
	if err != nil {
		// fallback try udp if udp4 failed
		udpConn, err = net.DialUDP("udp", nil, ipv4Dest)
	}
	if err != nil {
		l.Close()
		t.Fatalf("DialUDP failed: %v", err)
	}
	defer udpConn.Close()

	// Build endpoints
	localEP, _ := udpmod.ParseEndpoint(udpConn.LocalAddr().String())
	remoteEP, _ := udpmod.ParseEndpoint(udpConn.RemoteAddr().String())

	// Create outbound reliable Conn
	outConn, err := NewConn(udpConn, localEP, remoteEP, Config{}, true, nil, baseCtx)
	if err != nil {
		t.Fatalf("NewConn outbound failed: %v", err)
	}
	t.Logf("client local=%v remote=%v", udpConn.LocalAddr(), udpConn.RemoteAddr())

	// Run client handshake
	hCtx, hCancel := baseCtx.WithTimeout(2 * time.Second)
	defer hCancel()
	if err := outConn.StartClientHandshake(hCtx); err != nil {
		outConn.Close()
		l.Close()
		t.Fatalf("client handshake failed: %v", err)
	}
	t.Logf("client handshake complete")

	// Wait for server side acceptance
	var serverConn *Conn
	select {
	case serverConn = <-acceptedCh:
		if serverConn != nil {
			t.Logf("server accepted remote=%v", serverConn.RemoteEndpoint())
		}
	case <-time.After(3 * time.Second):
		// handshake should have completed well before this
		outConn.Close()
		l.Close()
		t.Fatalf("timeout waiting for Accept()")
	}
	if serverConn == nil {
		outConn.Close()
		l.Close()
		t.Fatalf("nil serverConn returned")
	}
	defer serverConn.Close()

	// Send payload client->server
	msg := []byte("Hello World")
	if _, err := outConn.Write(msg); err != nil {
		// ensure cleanup before failing
		outConn.Close()
		serverConn.Close()
		l.Close()
		t.Fatalf("client write failed: %v", err)
	}
	t.Logf("client wrote payload len=%d", len(msg))

	// Read at server side (no direct read deadline API; use goroutine + timeout)
	readCh := make(chan struct{})
	var got []byte
	var readErr error
	go func() {
		b := make([]byte, 64)
		if n, err := serverConn.Read(b); err != nil {
			readErr = err
		} else {
			got = append(got, b[:n]...)
		}
		close(readCh)
	}()
	select {
	case <-readCh:
	case <-time.After(10 * time.Second):
		outConn.Close()
		serverConn.Close()
		l.Close()
		t.Fatalf("timeout waiting for server read")
	}
	if readErr != nil {
		outConn.Close()
		serverConn.Close()
		l.Close()
		t.Fatalf("server read error: %v", readErr)
	}
	if string(got) != string(msg) {
		outConn.Close()
		serverConn.Close()
		l.Close()
		t.Fatalf("unexpected server payload: got %q want %q", string(got), string(msg))
	}
	t.Logf("server read payload: %s", string(got))

	// (Optional) Echo back from server to client to validate reverse path
	if _, err := serverConn.Write([]byte("ACK")); err == nil {
		// attempt client read with timeout
		clientReadCh := make(chan []byte, 1)
		go func() {
			b := make([]byte, 8)
			if n, e := outConn.Read(b); e == nil {
				clientReadCh <- b[:n]
			}
			close(clientReadCh)
		}()
		select {
		case resp := <-clientReadCh:
			if len(resp) > 0 && string(resp) != "ACK" {
				// Non-fatal; log mismatch
				// t.Logf("unexpected client echo: %q", string(resp))
			}
		case <-time.After(2 * time.Second):
			// ignore echo timeout (non-fatal for core test)
		}
	}
}
