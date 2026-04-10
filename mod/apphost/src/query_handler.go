package apphost

import (
	"errors"
	"fmt"
	"io"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/ipc"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/mod/apphost"
)

// QueryHandler is an astral.Router that routes queries to an IPC endpoint.
type QueryHandler struct {
	Identity *astral.Identity // identity of the handler
	IpcToken astral.Nonce     // token for the IPC endpoint
	Endpoint string           // IPC endpoint
}

var errEndpointUnavailable = errors.New("endpoint unavailable")

func (handler *QueryHandler) RouteQuery(ctx *astral.Context, q *astral.InFlightQuery, w io.WriteCloser) (io.WriteCloser, error) {
	conn, err := ipc.DialContext(ctx, handler.Endpoint)
	if err != nil {
		return nil, errEndpointUnavailable
	}

	var ch = channel.New(conn)

	err = ch.Send(&apphost.HandleQueryMsg{
		IpcToken: handler.IpcToken,
		ID:       q.Nonce,
		Caller:   q.Caller,
		Target:   q.Target,
		Query:    astral.String16(q.QueryString),
	})
	if err != nil {
		return query.RouteNotFound(handler, err)
	}

	err = ch.Switch(
		channel.ExpectAck,
		func(msg *apphost.QueryRejectedMsg) error {
			return astral.NewErrRejected(uint8(msg.Code))
		},
		func(msg *apphost.ErrorMsg) error {
			return astral.NewErrRouteNotFound(handler)
		},
		func(object astral.Object) error { // catch all
			return astral.NewErrRouteNotFound(handler, astral.NewErrUnexpectedObject(object))
		},
		func(err error) error {
			return astral.NewErrRouteNotFound(handler, fmt.Errorf("receive error: %v", err))
		},
	)
	if err != nil {
		return nil, err
	}

	// proxy the connection traffic
	go func() {
		io.Copy(w, conn)
		w.Close()
	}()

	return conn, nil
}
