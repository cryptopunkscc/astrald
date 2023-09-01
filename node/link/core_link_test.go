package link

import (
	"bytes"
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/net"
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

func (n *FakeConn) Outbound() bool {
	return n.outbound
}

func (n *FakeConn) LocalEndpoint() net.Endpoint {
	return net.NewGenericEndpoint("none", []byte{0})
}

func (n *FakeConn) RemoteEndpoint() net.Endpoint {
	return net.NewGenericEndpoint("none", []byte{0})
}

func (n *FakeConn) Close() error {
	return n.ReadWriteCloser.Close()
}

const msg = "IMPORTANT: hello from the other side\n"

func NewSecureConn(localIdentity id.Identity, remoteIdentity id.Identity, readWriteCloser io.ReadWriteCloser) *SecureConn {
	return &SecureConn{localIdentity: localIdentity, remoteIdentity: remoteIdentity, ReadWriteCloser: readWriteCloser}
}

func (s SecureConn) Outbound() bool               { return false }
func (s SecureConn) LocalEndpoint() net.Endpoint  { return nil }
func (s SecureConn) RemoteEndpoint() net.Endpoint { return nil }
func (s *SecureConn) RemoteIdentity() id.Identity { return s.remoteIdentity }
func (s *SecureConn) LocalIdentity() id.Identity  { return s.localIdentity }

type TestRouter struct {
	t *testing.T
}

func (l *TestRouter) RouteQuery(ctx context.Context, query net.Query, caller net.SecureWriteCloser, hints net.Hints) (net.SecureWriteCloser, error) {
	return net.Accept(query, caller, func(conn net.SecureConn) {
		_, err := conn.Write([]byte(msg))
		if err != nil {
			l.t.Fatal(err)
		}
		conn.Close()
	})
}

func TestCoreLink(t *testing.T) {
	var ctx, cancel = context.WithCancel(context.Background())
	var id1, _ = id.GenerateIdentity()
	var id2, _ = id.GenerateIdentity()

	var id1conn, id2conn = streams.Pipe()
	var id1link = NewCoreLink(NewSecureConn(id1, id2, id1conn))
	var id2link = NewCoreLink(NewSecureConn(id2, id1, id2conn))
	var wg sync.WaitGroup

	id1link.SetUplink(&TestRouter{t})
	id2link.SetUplink(&TestRouter{t})

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

		conn, err := net.Route(ctx, id2link, net.NewQuery(id2, id1, "testing"))
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

		link, err := Accept(ctx, conn, leftID)
		if err != nil {
			t.Fatal(err)
			return
		}
		link.Close()
	}()

	go func() {
		defer wg.Done()

		conn := &FakeConn{ReadWriteCloser: right}

		link, err := Open(ctx, conn, leftID, rightID)
		if err != nil {
			t.Fatal(err)
			return
		}
		link.Close()
	}()

	wg.Wait()
}
