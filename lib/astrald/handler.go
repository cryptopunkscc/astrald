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

type Handler struct {
	listener net.Listener
	ipcToken astral.Nonce
	doneCh   chan struct{}
	done     atomic.Bool
	client   *Client
}

// NewHandler creates a new Handler using apphost's DefaultRouter() protocol and a random auth ipcToken.
// Pass nil registrar for an IPC listener with no node registration.
func NewHandler(ctx *astral.Context, r apphost.Registrar) (*Handler, error) {
	return NewHandlerAt(ctx, Default(), libapphost.DefaultRouter().Protocol(), astral.NewNonce(), r)
}

// NewHandlerAt creates a new Handler for the given client, protocol and auth ipcToken.
// Pass nil registrar for an IPC listener with no node registration.
func NewHandlerAt(ctx *astral.Context, client *Client, protocol string, token astral.Nonce, registrar apphost.Registrar) (*Handler, error) {
	l, err := ipc.ListenAny(protocol)
	if err != nil {
		return nil, err
	}

	h := &Handler{
		listener: l,
		doneCh:   make(chan struct{}),
		ipcToken: token,
		client:   client,
	}

	if registrar == nil {
		return h, nil
	}

	if err = registrar.Register(ctx, h.Endpoint(), token); err != nil {
		h.Close()
		return nil, err
	}

	return h, nil
}

// ReadQuery waits for and returns the next pending query
func (h *Handler) ReadQuery() (*PendingQuery, error) {
	for {
		conn, err := h.listener.Accept()
		if err != nil {
			h.Close()
			return nil, err
		}
		ch := channel.New(conn)

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

		// check auth ipcToken
		if queryMsg.IpcToken != h.ipcToken {
			ch.Send(&apphost.ErrorMsg{Code: apphost.ErrCodeDenied})
			ch.Close()
			return nil, ErrInvalidToken
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

type HandleFunc func(ctx *astral.Context, query *PendingQuery) error

// Serve calls the given HandleFunc for every query received
func (h *Handler) Serve(ctx *astral.Context, handle HandleFunc) error {
	for {
		// get the next pending query
		pending, err := h.ReadQuery()
		if err != nil {
			return err
		}

		err = handle(ctx, pending)
		if err != nil {
			return err
		}
	}
}

// Route routes every query to given astral.Router
func (h *Handler) Route(ctx *astral.Context, router astral.Router) error {
	var errRejected *astral.ErrRejected

	for {
		// get the next pending query
		pending, err := h.ReadQuery()
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

// SetToken sets the token expected by the Handler
func (h *Handler) SetToken(token astral.Nonce) {
	h.ipcToken = token
}

// Token returns the token expected by the Handler
func (h *Handler) Token() astral.Nonce {
	return h.ipcToken
}

func (h *Handler) Close() error {
	if h.done.CompareAndSwap(false, true) {
		close(h.doneCh)
	}
	return h.listener.Close()
}

func (h *Handler) String() string {
	return h.Endpoint()
}

func (h *Handler) Endpoint() string {
	a := h.listener.Addr()
	return a.Network() + ":" + a.String()
}

// Done returns a channel that will be closed when the Handler is closed
func (h *Handler) Done() <-chan struct{} {
	return h.doneCh
}
