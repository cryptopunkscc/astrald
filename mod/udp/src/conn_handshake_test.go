package udp

import (
	"context"
	"testing"
	"time"
)

func TestClientHandshake_Success(t *testing.T) {
	conn := &Conn{
		notifyInbound: make(chan *Packet, 1),
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// Simulate server SYN|ACK response
	go func() {
		time.Sleep(10 * time.Millisecond)
		conn.notifyInbound <- &Packet{
			Seq:   12345, // server ISN
			Ack:   2,     // client ISN + 1
			Flags: FlagSYN | FlagACK,
			Len:   0,
		}
	}()

	conn.initialSeqNumLocal = 1 // deterministic for test
	conn.connID = 1

	err := conn.startClientHandshake(ctx)
	if err != nil {
		t.Fatalf("expected handshake success, got error: %v", err)
	}
	if conn.state != StateEstablished {
		t.Fatalf("expected StateEstablished, got %v", conn.state)
	}
	if conn.initialSeqNumRemote != 12345 {
		t.Fatalf("expected initialSeqNumRemote=12345, got %v", conn.initialSeqNumRemote)
	}
}

func TestClientHandshake_Timeout(t *testing.T) {
	conn := &Conn{
		notifyInbound: make(chan *Packet, 1),
	}
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := conn.startClientHandshake(ctx)
	if err == nil || err.Error() != "handshake timeout" {
		t.Fatalf("expected handshake timeout error, got: %v", err)
	}
}

func TestServerHandshake_Success(t *testing.T) {
	conn := &Conn{
		notifyInbound: make(chan *Packet, 1),
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// Simulate client ACK response
	go func() {
		time.Sleep(10 * time.Millisecond)
		conn.notifyInbound <- &Packet{
			Ack:   2, // server ISN + 1
			Flags: FlagACK,
			Len:   0,
		}
	}()

	synPkt := &Packet{Seq: 1, Flags: FlagSYN, Len: 0}
	err := conn.startServerHandshake(ctx, synPkt)
	if err != nil {
		t.Fatalf("expected handshake success, got error: %v", err)
	}
	if conn.state != StateEstablished {
		t.Fatalf("expected StateEstablished, got %v", conn.state)
	}
}

func TestServerHandshake_Timeout(t *testing.T) {
	conn := &Conn{
		notifyInbound: make(chan *Packet, 1),
	}
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	synPkt := &Packet{Seq: 1, Flags: FlagSYN, Len: 0}
	err := conn.startServerHandshake(ctx, synPkt)
	if err == nil || err.Error() != "handshake timeout" {
		t.Fatalf("expected handshake timeout error, got: %v", err)
	}
}

func TestClientHandshake_BadAckIgnored(t *testing.T) {
	conn := &Conn{
		notifyInbound: make(chan *Packet, 2),
	}
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Simulate server response with wrong Ack
	go func() {
		time.Sleep(10 * time.Millisecond)
		conn.notifyInbound <- &Packet{
			Seq:   12345,
			Ack:   999, // wrong ack
			Flags: FlagSYN | FlagACK,
			Len:   0,
		}
		// Then correct response
		time.Sleep(10 * time.Millisecond)
		conn.notifyInbound <- &Packet{
			Seq:   12345,
			Ack:   2,
			Flags: FlagSYN | FlagACK,
			Len:   0,
		}
	}()

	conn.initialSeqNumLocal = 1
	conn.connID = 1

	err := conn.startClientHandshake(ctx)
	if err != nil {
		t.Fatalf("expected handshake success, got error: %v", err)
	}
	if conn.state != StateEstablished {
		t.Fatalf("expected StateEstablished, got %v", conn.state)
	}
}
