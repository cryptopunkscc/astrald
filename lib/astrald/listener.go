package astrald

import (
	"errors"
	"io"
	"net"
	"sync/atomic"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	libapphost "github.com/cryptopunkscc/astrald/lib/apphost"
	"github.com/cryptopunkscc/astrald/lib/ipc"
	"github.com/cryptopunkscc/astrald/mod/apphost"
)

type Listener struct {
	net.Listener
	token  astral.Nonce
	doneCh chan struct{}
	done   atomic.Bool
}

var _ net.Listener = &Listener{}

// Listen creates a new Listener using apphost's DefaultRouter() protocol and a random auth token
func Listen() (*Listener, error) {
	return ListenAt(libapphost.DefaultRouter().Protocol(), astral.NewNonce())
}

// ListenAt creates a new Listener for the given protocol with the provided auth token
func ListenAt(protocol string, authToken astral.Nonce) (*Listener, error) {
	l, err := ipc.ListenAny(protocol)
	if err != nil {
		return nil, err
	}

	return &Listener{
		Listener: l,
		doneCh:   make(chan struct{}),
		token:    authToken,
	}, nil
}

// Next waits for and returns the next pending query
func (listener *Listener) Next() (*PendingQuery, error) {
	for {
		// accept the next network connection
		conn, err := listener.Listener.Accept()
		if err != nil {
			listener.Close()
			return nil, err
		}
		ch := channel.New(conn)

		// read the query request
		obj, err := ch.Receive()
		if err != nil {
			ch.Close()
			return nil, err
		}

		// check message type - must be HandleQueryMsg
		queryMsg, ok := obj.(*apphost.HandleQueryMsg)
		if !ok {
			ch.Send(&apphost.ErrorMsg{Code: apphost.ErrCodeProtocolError})
			ch.Close()
			continue
		}

		// check auth token
		if queryMsg.AuthToken != listener.token {
			ch.Send(&apphost.ErrorMsg{Code: apphost.ErrCodeDenied})
			ch.Close()
			return nil, ErrInvalidAuthToken
		}

		// return the pending query
		return &PendingQuery{
			conn: conn,
			query: &astral.Query{
				Nonce:  queryMsg.ID,
				Caller: queryMsg.Caller,
				Target: queryMsg.Target,
				Query:  string(queryMsg.Query),
			},
		}, nil
	}
}

// Serve forwards all incoming queries to the provided astral.Router
func (listener *Listener) Serve(ctx *astral.Context, router astral.Router) error {
	var errRejected *astral.ErrRejected

	for {
		// get the next pending query
		pending, err := listener.Next()
		if err != nil {
			return err
		}

		// lock the writer so that we can send the query response before the target starts sending data
		lockedWriter := NewLockableWriteCloser(pending.conn)
		lockedWriter.Lock()

		// route the query to the target
		w, err := router.RouteQuery(ctx, pending.query, lockedWriter)
		switch {
		case err == nil:
			// accepted - send an Ack and release the writer to the query target
			conn := pending.Accept()
			lockedWriter.Unlock()

			// forward the traffic
			go func() {
				io.Copy(w, conn)
				w.Close()
			}()

		case errors.As(err, &errRejected):
			// rejected - forward the rejection code and release the writer
			pending.RejectWithCode(int(errRejected.Code))
			lockedWriter.Unlock()

		case errors.Is(err, &astral.ErrRouteNotFound{}):
			// route not found - send route not found and release the writer
			pending.Skip()
			lockedWriter.Unlock()

		default:
			// unexpected error - skip the response and release the writer
			// TODO: should we have a different response to this than to route not found?
			pending.Skip()
			lockedWriter.Unlock()
		}
	}
}

// SetAuthToken sets the auth token expected by the Listener
func (listener *Listener) SetAuthToken(token astral.Nonce) {
	listener.token = token
}

// AuthToken returns the auth token expected by the Listener
func (listener *Listener) AuthToken() astral.Nonce {
	return listener.token
}

// Accept accepts the next pending query
func (listener *Listener) Accept() (net.Conn, error) {
	q, err := listener.Next()
	if err != nil {
		return nil, err
	}

	return q.Accept(), nil
}

func (listener *Listener) Close() error {
	if listener.done.CompareAndSwap(false, true) {
		close(listener.doneCh)
	}
	return listener.Listener.Close()
}

func (listener *Listener) Addr() net.Addr {
	return listener.Listener.Addr()
}

func (listener *Listener) String() string {
	return listener.Endpoint()
}

func (listener *Listener) Endpoint() string {
	a := listener.Listener.Addr()
	return a.Network() + ":" + a.String()
}

// Done returns a channel that will be closed when the Listener is closed
func (listener *Listener) Done() <-chan struct{} {
	return listener.doneCh
}
