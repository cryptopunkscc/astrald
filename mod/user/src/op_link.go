package user

import (
	"io"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/routing"
)

type opLinkArgs struct {
}

func (mod *Module) OpLink(ctx *astral.Context, q *routing.IncomingQuery, args opLinkArgs) (err error) {
	ac := mod.ActiveContract()
	if ac == nil {
		return q.Reject()
	}

	// only the current user can link
	if !q.Caller().IsEqual(ac.UserID) {
		return q.Reject()
	}

	conn := q.AcceptRaw()
	defer conn.Close()

	var done = make(chan struct{})
	defer close(done)

	go func() {
		select {
		case <-ctx.Done():
			conn.Close()
		case <-done:
		}
	}()

	io.Copy(io.Discard, conn)

	return nil
}
