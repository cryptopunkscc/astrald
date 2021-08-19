package proto

import (
	"encoding/json"
	"io"
)

type Conn struct {
	io.ReadWriteCloser
	enc *json.Encoder
	dec *json.Decoder
}

func NewConn(rwc io.ReadWriteCloser) *Conn {
	return &Conn{
		ReadWriteCloser: rwc,
		enc:             json.NewEncoder(rwc),
		dec:             json.NewDecoder(rwc),
	}
}

func (socket *Conn) ReadRequest() (req Request, err error) {
	err = socket.dec.Decode(&req)
	return
}

func (socket *Conn) ReadResponse() (res Response, err error) {
	err = socket.dec.Decode(&res)
	return
}

func (socket *Conn) Connect(identity string, port string) (res Response, err error) {
	err = socket.enc.Encode(&Request{
		Type:     RequestConnect,
		Identity: identity,
		Port:     port,
	})
	if err != nil {
		return
	}

	return socket.ReadResponse()
}

func (socket *Conn) Register(port string, path string) (res Response, err error) {
	err = socket.enc.Encode(&Request{
		Type: RequestRegister,
		Port: port,
		Path: path,
	})
	if err != nil {
		return
	}

	return socket.ReadResponse()
}

func (socket *Conn) Error(err string) error {
	socket.enc.Encode(&Response{
		Status: StatusError,
		Error:  err,
	})
	return socket.Close()
}

func (socket *Conn) OK() error {
	return socket.enc.Encode(&Response{
		Status: StatusOK,
	})
}
