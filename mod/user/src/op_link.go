package user

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/shell"
	"io"
)

type opLinkArgs struct {
}

func (mod *Module) OpLink(ctx *astral.Context, q shell.Query, args opLinkArgs) (err error) {
	ac := mod.ActiveContract()
	if ac == nil {
		return q.Reject()
	}

	// only the current user can link
	if !q.Caller().IsEqual(ac.UserID) {
		return q.Reject()
	}

	conn := q.Accept()
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
