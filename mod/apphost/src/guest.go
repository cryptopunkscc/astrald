package apphost

import (
	"errors"
	"io"
	"net"
	"strings"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/mod/apphost"
	"github.com/cryptopunkscc/astrald/mod/auth"
	"github.com/cryptopunkscc/astrald/streams"
)

// Guest represents an active connection with a guest.
type Guest struct {
	*channel.Channel
	mod     *Module
	guestID *astral.Identity
	conn    io.ReadWriteCloser
}

func NewGuest(mod *Module, conn net.Conn) *Guest {
	return &Guest{
		mod:     mod,
		conn:    conn,
		Channel: channel.New(conn),
	}
}

// Serve handles the guest connection.
func (guest *Guest) Serve(ctx *astral.Context) (err error) {
	// write a welcome message
	err = guest.Write(&apphost.HostInfoMsg{
		Identity: ctx.Identity(),
		Alias:    astral.String8(guest.mod.Dir.DisplayName(ctx.Identity())),
	})
	if err != nil {
		return
	}

	// message read loop
	var msg astral.Object
	for {
		msg, err = guest.Read()

		// check err
		switch {
		case err == nil:
		case errors.Is(err, io.EOF):
			return nil
		case strings.Contains(err.Error(), "use of closed network connection"):
			return nil
		default:
			guest.mod.log.Logv(2, "error reading from client: %v", err)
			return err
		}

		// handle message
		switch msg := msg.(type) {
		case *apphost.AuthTokenMsg:
			err = guest.onAuthTokenMsg(ctx, msg)
		case *apphost.RegisterHandlerMsg:
			err = guest.onRegisterHandlerMsg(ctx, msg)
		case *apphost.RouteQueryMsg:
			err = guest.onRouteQueryMsg(ctx, msg)
		case *apphost.PingMsg:
			err = guest.Write(&astral.Ack{})
		default:
			guest.mod.log.Logv(1, "protocol error: invalid message: %v", msg.ObjectType())
			return guest.Write(&apphost.ErrorMsg{Code: apphost.ErrCodeProtocolError})
		}

		if err != nil {
			return
		}
	}
}

func (guest *Guest) onAuthTokenMsg(ctx *astral.Context, msg *apphost.AuthTokenMsg) (err error) {
	// fetch info about the token from the database
	dbToken, err := guest.mod.db.FindAccessToken(string(msg.Token))

	if dbToken == nil {
		guest.mod.log.Errorv(3, "token authentication failed")
		return guest.Write(&apphost.ErrorMsg{Code: apphost.ErrCodeAuthFailed})
	}

	guest.guestID = dbToken.Identity
	guest.mod.log.Infov(3, "%v authenticated using auth token", guest.guestID)

	return guest.Write(&apphost.AuthSuccessMsg{
		GuestID: guest.guestID,
	})
}

func (guest *Guest) onRegisterHandlerMsg(ctx *astral.Context, msg *apphost.RegisterHandlerMsg) (err error) {
	// only authenticated guests can register handlers
	if !guest.isAuthenticated() {
		return guest.Write(&apphost.ErrorMsg{Code: apphost.ErrCodeDenied})
	}

	// if requested identity is different from the authenticated identity, check authorization
	if !msg.Identity.IsEqual(guest.guestID) {
		if !guest.mod.Auth.Authorize(guest.guestID, auth.ActionSudo, msg.Identity) {
			return guest.Write(&apphost.ErrorMsg{Code: apphost.ErrCodeDenied})
		}
	}

	// add the handler
	handler := &QueryHandler{
		Identity:  msg.Identity,
		AuthToken: msg.AuthToken,
		Endpoint:  string(msg.Endpoint),
	}
	guest.mod.handlers.Add(handler)
	defer guest.mod.handlers.Remove(handler) // remove handler on disconnect

	// send ack to the client
	err = guest.Write(&astral.Ack{})
	if err != nil {
		return
	}

	defer guest.Close()
	guest.mod.log.Logv(3, "%v registered a handler for %v at %v", guest.guestID, handler.Identity, handler.Endpoint)

	// NOTE: at this stage this guest connection is only used to keep the query handler alive

	// close connection if context ends
	var done = make(chan struct{})
	defer close(done)
	go func() {
		select {
		case <-ctx.Done():
			guest.Close()
		case <-done:
		}
	}()

	// wait for the connection to end ignoring all incoming objects
	for {
		_, err = guest.Read()
		if err != nil {
			break
		}
	}

	return nil
}

func (guest *Guest) onRouteQueryMsg(ctx *astral.Context, msg *apphost.RouteQueryMsg) (err error) {
	// deny if not authenticated and anonymous queries are not allowed
	if !guest.isAuthenticated() && !guest.mod.config.AllowAnonymous {
		return guest.Write(&apphost.ErrorMsg{Code: apphost.ErrCodeDenied})
	}

	var q = &astral.Query{
		Nonce:  msg.Nonce,
		Caller: msg.Caller,
		Target: msg.Target,
		Query:  string(msg.Query),
	}

	// check authorization if necessary
	switch {
	case q.Caller.IsZero():
	case q.Caller.IsEqual(guest.guestID):
	default:
		if !guest.mod.Authorize(guest.guestID, auth.ActionSudo, q.Caller) {
			return guest.Write(&apphost.ErrorMsg{Code: apphost.ErrCodeDenied})
		}
	}

	// set network access depending on authentication
	if !guest.isAuthenticated() {
		ctx = ctx.ExcludeZone(astral.ZoneNetwork)
	} else {
		ctx = ctx.IncludeZone(astral.ZoneNetwork)
	}

	// route the query
	conn, err := query.Route(ctx, guest.mod.node, q)

	// check error
	var rejected *astral.ErrRejected
	switch {
	case err == nil:
	case errors.As(err, &rejected):
		return guest.Write(&apphost.QueryRejectedMsg{Code: astral.Uint8(rejected.Code)})
	case errors.Is(err, &astral.ErrRouteNotFound{}):
		return guest.Write(&apphost.ErrorMsg{Code: apphost.ErrCodeRouteNotFound})
	default:
		return guest.Write(&apphost.ErrorMsg{Code: apphost.ErrCodeInternalError})
	}

	// write success response
	err = guest.Write(&apphost.QueryAcceptedMsg{})
	if err != nil {
		return err
	}

	// proxy the traffic
	_, _, err = streams.Join(guest.conn, conn)

	return
}

func (guest *Guest) isAuthenticated() bool {
	return !guest.guestID.IsZero()
}
