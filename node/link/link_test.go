package link

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/infra"
	"github.com/cryptopunkscc/astrald/streams"
	"io"
	"sync"
	"testing"
)

type MockAuth struct {
	io.ReadWriteCloser
	outbound bool
}

func (conn *MockAuth) Outbound() bool {
	return conn.outbound
}

func (conn *MockAuth) LocalAddr() infra.Addr {
	return infra.NewGenericAddr("test", []byte{})
}

func (conn *MockAuth) RemoteAddr() infra.Addr {
	return infra.NewGenericAddr("test", []byte{})
}

func (conn *MockAuth) RemoteIdentity() id.Identity {
	return id.Identity{}
}

func (conn *MockAuth) LocalIdentity() id.Identity {
	return id.Identity{}
}

func TestLink(t *testing.T) {
	var s, c = streams.Pipe()
	var ctx = context.Background()
	var wg = sync.WaitGroup{}

	// server
	wg.Add(1)
	go func() {
		defer wg.Done()
		var link = New(&MockAuth{ReadWriteCloser: s})

		link.SetQueryHandler(func(query *Query) error {
			return query.Reject()
		})

		if err := link.Run(ctx); !errors.Is(err, io.EOF) {
			t.Fatal(err)
		}
	}()

	// client
	wg.Add(1)
	go func() {
		defer wg.Done()
		var link = New(&MockAuth{ReadWriteCloser: c})

		link.SetQueryHandler(func(query *Query) error {
			return query.Reject()
		})

		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := link.Run(ctx); !errors.Is(err, io.EOF) {
				t.Fatal(err)
			}
		}()

		if _, err := link.Query(ctx, "test-reject"); err != ErrRejected {
			t.Fatal("unexpected error:", err)
		}

		link.Close()
	}()

	wg.Wait()
}
