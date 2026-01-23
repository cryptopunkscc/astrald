package astrald

import (
	"errors"
	"io"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/apphost"
)

type Server struct {
	astral.Router
	*Listener
}

func NewServer(listener *Listener, router astral.Router) *Server {
	return &Server{
		Router:   router,
		Listener: listener,
	}
}

func (s *Server) Serve(ctx *astral.Context) error {
	var errRejected *astral.ErrRejected

	for {
		q, err := s.Next()
		if err != nil {
			return err
		}

		var conn *apphost.Conn

		w, err := s.RouteQuery(ctx, q.query, q.conn)
		switch {
		case err == nil:
			conn = q.Accept()

		case errors.As(err, &errRejected):
			q.RejectWithCode(int(errRejected.Code))
			continue

		default:
			q.Reject()
			continue
		}

		go func() {
			io.Copy(w, conn)
			w.Close()
		}()
	}
}
