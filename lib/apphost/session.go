package apphost

import (
	"errors"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/ipc"
	"github.com/cryptopunkscc/astrald/mod/apphost"
	"net"
)

type Session struct {
	conn net.Conn
	addr string
}

func Connect(addr string) (*Session, error) {
	conn, err := ipc.Dial(addr)
	if err != nil {
		return nil, err
	}

	return &Session{
		conn: conn,
		addr: addr,
	}, nil
}

func (s *Session) Token(token string) (res apphost.TokenResponse, err error) {
	// write method name
	_, err = (*astral.String8)(astral.NewString("token")).WriteTo(s.conn)
	if err != nil {
		return
	}

	// write args
	_, err = (*astral.String8)(&token).WriteTo(s.conn)
	if err != nil {
		return
	}

	// read response
	_, err = res.ReadFrom(s.conn)
	return
}

func (s *Session) Query(callerID *astral.Identity, targetID *astral.Identity, query string) (conn *Conn, err error) {
	// write method name
	_, err = (*astral.String8)(astral.NewString("query")).WriteTo(s.conn)
	if err != nil {
		return
	}

	// write args
	_, err = apphost.QueryArgs{
		Caller: callerID,
		Target: targetID,
		Query:  astral.String16(query),
	}.WriteTo(s.conn)
	if err != nil {
		return
	}

	var code astral.Uint8
	_, err = code.ReadFrom(s.conn)
	if err != nil {
		return
	}

	if code != 0 {
		return nil, &astral.ErrRejected{Code: uint8(code)}
	}

	return &Conn{
		Conn:     s.conn,
		remoteID: targetID,
		query:    query,
	}, nil
}

func (s *Session) Register(identity *astral.Identity, target string) (token string, err error) {
	// write method name
	_, err = (*astral.String8)(astral.NewString("register")).WriteTo(s.conn)
	if err != nil {
		return
	}

	// write args
	_, err = apphost.RegisterArgs{
		Identity: identity,
		Endpoint: astral.String8(target),
		Flags:    0,
	}.WriteTo(s.conn)

	var code astral.Uint8
	_, err = code.ReadFrom(s.conn)

	switch code {
	case 0:
	case 1:
		return "", errors.New("unauthorized")
	case 2:
		return "", errors.New("already registered")
	default:
		return "", errors.New("unknown error")
	}

	_, err = (*astral.String8)(&token).ReadFrom(s.conn)

	return
}

func (s *Session) Close() error {
	return s.conn.Close()
}
