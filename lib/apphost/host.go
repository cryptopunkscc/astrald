package apphost

import (
	"fmt"
	"net"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/ipc"
	"github.com/cryptopunkscc/astrald/mod/apphost"
)

type Host struct {
	*channel.Channel
	conn      net.Conn
	hostID    *astral.Identity
	hostAlias string
	guestID   *astral.Identity
}

// Connect connects to the apphost endpoint.
func Connect(endpoint string) (*Host, error) {
	conn, err := ipc.Dial(endpoint)
	if err != nil {
		return nil, err
	}
	ch := channel.New(conn)

	msg, err := ch.Receive()
	switch msg := msg.(type) {
	case *apphost.HostInfoMsg: // success
		return &Host{
			Channel:   ch,
			conn:      conn,
			hostID:    msg.Identity,
			hostAlias: string(msg.Alias),
		}, nil

	case nil: // error
		conn.Close()
		return nil, err

	default: // unexpected
		conn.Close()
		return nil, fmt.Errorf("unexpected message type: %s", msg.ObjectType())
	}
}

// AuthToken authenticates the session with the given auth token.
func (s *Host) AuthToken(token string) (err error) {
	// write auth request
	err = s.Send(&apphost.AuthTokenMsg{
		Token: astral.String8(token),
	})
	if err != nil {
		return err
	}

	// read response
	msg, err := s.Receive()
	switch msg := msg.(type) {
	case *apphost.AuthSuccessMsg:
		s.guestID = msg.GuestID
		return nil

	case *apphost.ErrorMsg:
		return msg

	case nil:
		return err

	default:
		return apphost.ErrProtocolError
	}
}

// RouteQuery routes a query via the host.
// If the caller is nil, the guest's identity is used. If the target is nil, the host's identity is used.
// If routing fails, the connection with the host is closed.
func (s *Host) RouteQuery(q *astral.Query, zone astral.Zone, filters []string) (conn *Conn, err error) {
	// close host connection on error
	defer func() {
		if conn == nil {
			s.Close()
		}
	}()

	// set the default caller
	if q.Caller == nil {
		q.Caller = s.GuestID()
	}

	// set the default target
	if q.Target == nil {
		q.Target = s.HostID()
	}

	var filters8 []astral.String8
	for _, f := range filters {
		filters8 = append(filters8, astral.String8(f))
	}

	// send query request
	err = s.Send(&apphost.RouteQueryMsg{
		Nonce:   q.Nonce,
		Caller:  q.Caller,
		Target:  q.Target,
		Query:   astral.String16(q.Query),
		Zone:    zone,
		Filters: filters8,
	})
	if err != nil {
		return
	}

	// handle response
	msg, err := s.Receive()
	switch msg := msg.(type) {
	case *apphost.QueryAcceptedMsg: // success
		return NewConn(s.conn, q, true), nil

	case *apphost.QueryRejectedMsg: // reject
		return nil, &astral.ErrRejected{Code: uint8(msg.Code)}

	case *apphost.ErrorMsg: // error
		return nil, msg

	case nil: // error
		return nil, err

	default: // unexpected
		return nil, apphost.ErrProtocolError
	}
}

// Register registers a query handler for the given identity. Token is an access token
// the host will to authenticate IPC calls. Close the host connection to unregister the handler.
func (s *Host) Register(identity *astral.Identity, target string, token astral.Nonce) (err error) {
	if identity.IsZero() {
		identity = s.GuestID()
	}

	err = s.Send(&apphost.RegisterHandlerMsg{
		Identity:  identity,
		Endpoint:  astral.String8(target),
		AuthToken: token,
	})
	if err != nil {
		return
	}

	// read response
	msg, err := s.Receive()
	switch msg := msg.(type) {
	case *astral.Ack:
		return nil

	case *apphost.ErrorMsg:
		return msg

	case nil:
		return err

	default:
		return apphost.ErrProtocolError
	}
}

// HostID returns the host identity.
func (s *Host) HostID() *astral.Identity {
	return s.hostID
}

// GuestID returns the guest identity.
func (s *Host) GuestID() *astral.Identity {
	return s.guestID
}

// HostAlias returns the host alias.
func (s *Host) HostAlias() string {
	return s.hostAlias
}

// Close closes the connection with the host.
func (s *Host) Close() error {
	return s.conn.Close()
}
