package apphost

import (
	"io"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/ipc"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/mod/apphost"
)

// QueryHandler is an astral.Router that routes queries to an IPC endpoint.
type QueryHandler struct {
	Identity  *astral.Identity // identity of the handler
	AuthToken astral.Nonce     // auth token for the IPC endpoint
	Endpoint  string           // IPC endpoint
}

func (handler *QueryHandler) RouteQuery(ctx *astral.Context, q *astral.Query, w io.WriteCloser) (io.WriteCloser, error) {
	conn, err := ipc.Dial(handler.Endpoint)
	if err != nil {
		return query.Reject()
	}

	var ch = channel.New(conn)

	err = ch.Send(&apphost.HandleQueryMsg{
		AuthToken: handler.AuthToken,
		ID:        q.Nonce,
		Caller:    q.Caller,
		Target:    q.Target,
		Query:     astral.String16(q.Query),
	})
	if err != nil {
		return query.Reject()
	}

	// read response
	msg, err := ch.Receive()
	switch msg := msg.(type) {
	case *astral.Ack: // success

	case *apphost.QueryRejectedMsg:
		return query.RejectWithCode(uint8(msg.Code))

	case *apphost.ErrorMsg:
		return query.RouteNotFound(handler, msg)

	case nil:
		return query.RouteNotFound(handler, err)

	default:
		return query.RouteNotFound(handler)
	}

	// proxy the connection traffic
	go func() {
		io.Copy(w, conn)
		w.Close()
	}()

	return conn, nil
}
