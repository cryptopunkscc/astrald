package muxlink

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"github.com/cryptopunkscc/astrald/id"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/streams"
	"io"
	"sync"
	"testing"
	"time"
)

type SecureConn struct {
	localIdentity  id.Identity
	remoteIdentity id.Identity
	io.ReadWriteCloser
}

type FakeConn struct {
	io.ReadWriteCloser
	outbound bool
}

type GenericEndpoint struct {
	network string
	bytes   []byte
}

func NewGenericEndpoint(network string, bytes []byte) *GenericEndpoint {
	return &GenericEndpoint{network: network, bytes: bytes}
}

func (g *GenericEndpoint) Network() string {
	return g.network
}

func (g *GenericEndpoint) Address() string {
	return hex.EncodeToString(g.bytes)
}

func (g *GenericEndpoint) Pack() []byte {
	return g.bytes
}

func (n *FakeConn) Outbound() bool {
	return n.outbound
}

func (n *FakeConn) LocalEndpoint() exonet.Endpoint {
	return NewGenericEndpoint("none", []byte{0})
}

func (n *FakeConn) RemoteEndpoint() exonet.Endpoint {
	return NewGenericEndpoint("none", []byte{0})
}

func (n *FakeConn) Close() error {
	return n.ReadWriteCloser.Close()
}

const msg = "IMPORTANT: hello from the other side\n"

func NewSecureConn(localIdentity id.Identity, remoteIdentity id.Identity, readWriteCloser io.ReadWriteCloser) *SecureConn {
	return &SecureConn{localIdentity: localIdentity, remoteIdentity: remoteIdentity, ReadWriteCloser: readWriteCloser}
}

func (s SecureConn) Outbound() bool                  { return false }
func (s SecureConn) LocalEndpoint() exonet.Endpoint  { return nil }
func (s SecureConn) RemoteEndpoint() exonet.Endpoint { return nil }
func (s *SecureConn) RemoteIdentity() id.Identity    { return s.remoteIdentity }
func (s *SecureConn) LocalIdentity() id.Identity     { return s.localIdentity }

type TestRouter struct {
	t *testing.T
}

func (l *TestRouter) RouteQuery(ctx context.Context, query astral.Query, caller astral.SecureWriteCloser, hints astral.Hints) (astral.SecureWriteCloser, error) {
	return astral.Accept(query, caller, func(conn astral.Conn) {
		_, err := conn.Write([]byte(msg))
		if err != nil {
			l.t.Fatal(err)
		}
		conn.Close()
	})
}

func TestLink(t *testing.T) {
	var ctx, cancel = context.WithCancel(context.Background())
	var id1, _ = id.GenerateIdentity()
	var id2, _ = id.GenerateIdentity()

	var id1conn, id2conn = streams.Pipe()
	var id1link = NewLink(NewSecureConn(id1, id2, id1conn), nil)
	var id2link = NewLink(NewSecureConn(id2, id1, id2conn), nil)
	var wg sync.WaitGroup

	id1link.SetLocalRouter(&TestRouter{t})
	id2link.SetLocalRouter(&TestRouter{t})

	wg.Add(1)
	go func() {
		defer wg.Done()
		id1link.Run(ctx)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		id2link.Run(ctx)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer cancel()

		if _, err := id1link.control.Ping(); err != nil {
			t.Fatal(err)
			return
		}

		if _, err := id2link.control.Ping(); err != nil {
			t.Fatal(err)
			return
		}

		conn, err := astral.Route(ctx, id2link, astral.NewQuery(id2, id1, "testing"))
		if err != nil {
			t.Fatal(err)
			return
		}

		var buf = &bytes.Buffer{}

		io.Copy(buf, conn)

		if bytes.Compare(buf.Bytes(), []byte(msg)) != 0 {
			t.Fatalf("received '%s', expected '%s'", string(buf.Bytes()), msg)
		}

		id2link.Close()

		// let links close due to EOF before canceling the context, so that we get proper errors
		time.Sleep(time.Millisecond)
	}()

	wg.Wait()

	if !errors.Is(id1link.Err(), ErrLinkClosedByPeer) {
		t.Fatalf("link1 closed with %s, expected %s", id1link.Err(), ErrLinkClosedByPeer)
	}
	if !errors.Is(id2link.Err(), ErrLinkClosed) {
		t.Fatalf("link2 closed with %s, expected %s", id1link.Err(), ErrLinkClosed)
	}
}

func TestOpenAccept(t *testing.T) {
	var wg sync.WaitGroup
	var left, right = streams.Pipe()
	var leftID, _ = id.GenerateIdentity()
	var rightID, _ = id.GenerateIdentity()
	var ctx = context.Background()

	wg.Add(2)
	go func() {
		defer wg.Done()

		conn := &FakeConn{ReadWriteCloser: left}

		link, err := Accept(ctx, conn, leftID, nil)
		if err != nil {
			t.Fatal(err)
			return
		}
		link.Close()
	}()

	go func() {
		defer wg.Done()

		conn := &FakeConn{ReadWriteCloser: right}

		link, err := Open(ctx, conn, leftID, rightID, nil)
		if err != nil {
			t.Fatal(err)
			return
		}
		link.Close()
	}()

	wg.Wait()
}
