package proto

import (
	"encoding/gob"
	"io"
)

type Socket struct {
	io.ReadWriteCloser
	enc *gob.Encoder
	dec *gob.Decoder
}

func NewSocket(rwc io.ReadWriteCloser) *Socket {
	return &Socket{
		ReadWriteCloser: rwc,
		enc:             gob.NewEncoder(rwc),
		dec:             gob.NewDecoder(rwc),
	}
}

func (socket *Socket) ReadRequest() (req Request, err error) {
	err = socket.dec.Decode(&req)
	return
}

func (socket *Socket) ReadResponse() (res Response, err error) {
	err = socket.dec.Decode(&res)
	return
}

func (socket *Socket) Connect(identity string, port string) (res Response, err error) {
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

func (socket *Socket) Register(port string, path string) (res Response, err error) {
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

func (socket *Socket) Error(err string) error {
	socket.enc.Encode(&Response{
		Status: StatusError,
		Error:  err,
	})
	return socket.Close()
}

func (socket *Socket) OK() error {
	return socket.enc.Encode(&Response{
		Status: StatusOK,
	})
}
