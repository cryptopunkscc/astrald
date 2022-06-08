package apphost

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/cslq"
	"io"
)

var _ AppHost = &Client{}

type Client struct {
	*cslq.Endec
	conn io.ReadWriteCloser
}

func Bind(conn io.ReadWriteCloser) *Client {
	return &Client{conn: conn, Endec: cslq.NewEndec(conn)}
}

func (client Client) Register(port string, target string) (err error) {
	var errorCode int

	// request
	if err = client.Encode("[c]c [c]c [c]c", cmdRegister, port, target); err != nil {
		return
	}

	// request
	if err = client.Decode("c", &errorCode); err != nil {
		return
	}

	// decode the error
	switch errorCode {
	case success:

	case errFailed:
		err = ErrFailed

	default:
		err = ErrInvalidErrorCode
	}
	return
}

func (client Client) Query(identity id.Identity, query string) (conn io.ReadWriteCloser, err error) {
	var errorCode int

	// request
	if err = cslq.Encode(client.conn, "[c]c v [c]c", cmdQuery, identity, query); err != nil {
		return
	}

	// request
	if err = cslq.Decode(client.conn, "c", &errorCode); err != nil {
		return
	}

	// decode the error
	switch errorCode {
	case success:
		conn = client.conn

	case errRejected:
		err = ErrRejected

	default:
		err = ErrInvalidErrorCode
	}

	return
}

func (client Client) Resolve(s string) (identity id.Identity, err error) {
	var errorCode int

	// request
	if err = cslq.Encode(client.conn, "[c]c [c]c", cmdResolve, s); err != nil {
		return
	}

	// response
	if err = cslq.Decode(client.conn, "c", &errorCode); err != nil {
		return
	}

	// decode the error
	switch errorCode {
	case success:
		err = cslq.Decode(client.conn, "v", &identity)

	case errFailed:
		err = ErrFailed

	default:
		err = ErrInvalidErrorCode
	}

	return
}

func (client Client) NodeInfo(identity id.Identity) (info NodeInfo, err error) {
	var errorCode int

	// request
	if err = cslq.Encode(client.conn, "[c]c v", cmdNodeInfo, identity); err != nil {
		return
	}

	// response
	if err = cslq.Decode(client.conn, "c v", &errorCode, &info); err != nil {
		return
	}

	// decode the error
	switch errorCode {
	case success:

	case errFailed:
		err = ErrFailed

	default:
		err = ErrInvalidErrorCode
	}

	return
}
