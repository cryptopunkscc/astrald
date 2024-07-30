package nodes

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/id"
	"github.com/cryptopunkscc/astrald/mod/nodes/src/frames"
	"io"
)

func (mod *Module) RouteQuery(ctx context.Context, query *astral.Query, caller io.WriteCloser) (w io.WriteCloser, err error) {
	if !mod.isRoutable(query.Target) {
		return astral.RouteNotFound(mod)
	}

	conn, ok := mod.conns.Set(query.Nonce, newConn(query.Nonce))
	if !ok {
		return astral.RouteNotFound(mod, errors.New("nonce already exists"))
	}

	conn.RemoteIdentity = query.Target
	conn.Query = query.Query
	conn.Outbound = true

	// make sure we're linked with the target node
	if err := mod.ensureConnected(ctx, query.Target); err != nil {
		conn.swapState(stateRouting, stateClosed)
		return astral.RouteNotFound(mod, err)
	}

	// prepare the protocol frame
	frame := &frames.Query{
		Nonce:  query.Nonce,
		Query:  query.Query,
		Buffer: uint32(conn.rsize),
	}

	// send the query via all streams
	for _, s := range mod.streams.Select(func(s *Stream) bool {
		return s.RemoteIdentity().IsEqual(query.Target)
	}) {
		go s.Write(frame)
	}

	// wait for the response
	select {
	case accepted := <-conn.res:
		if !accepted {
			return astral.Reject()
		}

		go func() {
			io.Copy(caller, conn)
			caller.Close()
		}()

		return conn, nil

	case <-ctx.Done():
		conn.swapState(stateRouting, stateClosed)
		return astral.RouteNotFound(mod, ctx.Err())
	}
}

func (mod *Module) isRoutable(identity id.Identity) bool {
	return mod.isLinked(identity) || mod.hasEndpoints(identity)
}
