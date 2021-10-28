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

func (conn *Conn) Query(nodeID string, query string) (res Response, err error) {
	err = writeJSON(conn, &Request{
		Type:     RequestQuery,
		Identity: nodeID,
		Port:     query,
	})
	if err != nil {
		return
	}

	return conn.ReadResponse()
}

func (conn *Conn) Register(name string, target string) (res Response, err error) {
	err = writeJSON(conn, &Request{
		Type: RequestRegister,
		Port: name,
		Path: target,
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
