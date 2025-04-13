package apphost

import (
	"errors"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/mod/apphost"
	"github.com/cryptopunkscc/astrald/streams"
	"io"
	"net"
)

type Session struct {
	mod     *Module
	guestID *astral.Identity
	conn    io.ReadWriteCloser
	log     *log.Logger
	bp      *astral.Blueprints
}

func NewSession(mod *Module, conn net.Conn, log *log.Logger) *Session {
	return &Session{
		mod:  mod,
		log:  log,
		conn: conn,
	}
}

func (s *Session) Serve(ctx *astral.Context) (err error) {
	var cmd astral.String8

	for {
		_, err = cmd.ReadFrom(s.conn)
		if err != nil {
			return
		}

		switch cmd {
		case "token":
			err = s.Token(ctx)

		case "anon":
			err = s.Anon(ctx)

		case "register":
			err = s.Register(ctx)

		case "query":
			err = s.Query(ctx)

		default:
			err = errors.New("invalid command")
		}

		if err != nil {
			return
		}
	}
}

// Anon starts an anonymous session
func (s *Session) Anon(ctx *astral.Context) (err error) {
	var res apphost.AuthResponse

	if !s.mod.config.AllowAnonymous {
		res = apphost.AuthResponse{
			Code:    apphost.Rejected,
			GuestID: nil,
			HostID:  nil,
		}
	} else {
		s.guestID = astral.Anyone

		res = apphost.AuthResponse{
			Code:    0,
			GuestID: astral.Anyone,
			HostID:  ctx.Identity(),
		}
	}

	_, err = res.WriteTo(s.conn)
	return
}

func (s *Session) Token(ctx *astral.Context) (err error) {
	var (
		arg apphost.TokenArgs
		res apphost.AuthResponse
	)

	_, err = arg.ReadFrom(s.conn)
	if err != nil {
		return
	}

	at, err := s.mod.db.FindAccessToken(string(arg.Token))

	if at == nil {
		s.log.Errorv(1, "token authentication failed")

		res = apphost.AuthResponse{
			Code:    apphost.Rejected,
			GuestID: nil,
			HostID:  nil,
		}
	} else {
		s.log.Infov(3, "authenticated as %v using a token", at.Identity)

		s.guestID = at.Identity

		res = apphost.AuthResponse{
			Code:    0,
			GuestID: at.Identity,
			HostID:  ctx.Identity(),
		}
	}

	_, err = res.WriteTo(s.conn)
	return
}

func (s *Session) Register(ctx *astral.Context) (err error) {
	var (
		arg apphost.RegisterArgs
	)

	_, err = arg.ReadFrom(s.conn)
	if err != nil {
		return
	}

	guestID := s.guestID
	if !arg.Identity.IsZero() {
		if s.mod.Auth.Authorize(guestID, admin.ActionSudo, arg.Identity) {
			guestID = s.guestID
		}
	}

	if guestID.IsZero() {
		_, err = apphost.RegisterResponse{
			Code:  apphost.Rejected,
			Token: "",
		}.WriteTo(s.conn)
		return
	}

	guest := &Guest{
		Token:    randomString(32),
		Identity: guestID,
		Endpoint: string(arg.Endpoint),
	}

	guestIDHex := guestID.String()

	_, ok := s.mod.guests.Set(guestIDHex, guest)
	if !ok {
		_, err = apphost.RegisterResponse{
			Code:  apphost.AlreadyRegistered,
			Token: "",
		}.WriteTo(s.conn)
		return
	}

	_, err = apphost.RegisterResponse{
		Code:  0,
		Token: astral.String8(guest.Token),
	}.WriteTo(s.conn)

	var done = make(chan struct{})
	defer s.conn.Close()

	go func() {
		select {
		case <-ctx.Done():
			s.conn.Close()
		case <-done:
		}
	}()

	s.log.Infov(1, "%v registered query handler %v", guest.Identity, guest.Endpoint)

	// wait for the connection to close (any data is a protocol violation)
	var p [1]byte
	s.conn.Read(p[:])

	s.mod.guests.Delete(guestIDHex)

	close(done)
	return errors.New("session ended")
}

func (s *Session) Query(ctx *astral.Context) (err error) {
	var (
		arg apphost.QueryArgs
	)

	_, err = arg.ReadFrom(s.conn)
	if err != nil {
		return
	}

	caller := s.guestID

	if caller.IsZero() && !s.mod.config.AllowAnonymous {
		_, err = apphost.QueryResponse{
			Code: apphost.Rejected,
		}.WriteTo(s.conn)
		return
	}

	if !arg.Caller.IsZero() && !arg.Caller.IsEqual(caller) {
		if !s.mod.Auth.Authorize(caller, admin.ActionSudo, arg.Caller) {
			_, err = apphost.QueryResponse{
				Code: apphost.Rejected,
			}.WriteTo(s.conn)
			return
		}
		caller = arg.Caller
	}

	var q = astral.NewQuery(caller, arg.Target, string(arg.Query))

	conn, err := query.Route(ctx, s.mod.node, q)

	if err != nil {
		var code = apphost.Rejected

		var r *astral.ErrRejected
		if errors.As(err, &r) {
			code = int(r.Code)
		}

		// write error response
		_, err = apphost.QueryResponse{Code: astral.Uint8(code)}.WriteTo(s.conn)
		return
	}

	// write success response
	_, err = apphost.QueryResponse{Code: apphost.Success}.WriteTo(s.conn)
	if err != nil {
		return
	}

	_, _, err = streams.Join(s.conn, conn)
	return
}
