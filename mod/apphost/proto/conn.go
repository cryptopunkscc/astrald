package proto

import (
	"io"
)

type Conn struct {
	io.ReadWriteCloser
}

func NewConn(rwc io.ReadWriteCloser) *Conn {
	return &Conn{
		ReadWriteCloser: rwc,
	}
}

func (conn *Conn) ReadRequest() (req Request, err error) {
	err = readJSON(conn, &req)
	return
}

func (conn *Conn) ReadResponse() (res Response, err error) {
	err = readJSON(conn, &res)
	return
}

func (conn *Conn) Connect(identity string, port string) (res Response, err error) {
	err = writeJSON(conn, &Request{
		Type:     RequestConnect,
		Identity: identity,
		Port:     port,
	})
	if err != nil {
		return
	}

	return conn.ReadResponse()
}

func (conn *Conn) Register(port string, path string) (res Response, err error) {
	err = writeJSON(conn, &Request{
		Type: RequestRegister,
		Port: port,
		Path: path,
	})
	if err != nil {
		return
	}

	return conn.ReadResponse()
}

func (conn *Conn) Error(errMsg string) error {
	err := writeJSON(conn, &Response{
		Status: StatusError,
		Error:  errMsg,
	})
	if err != nil {
		return err
	}
	return conn.Close()
}

func (conn *Conn) OK() error {
	return writeJSON(conn, &Response{
		Status: StatusOK,
	})
}
