package ctl

import (
	"errors"
	"github.com/cryptopunkscc/astrald/cslq"
	"io"
)

type Control struct {
	rw io.ReadWriter
}

func New(rw io.ReadWriter) *Control {
	return &Control{rw: rw}
}

func (ctl *Control) ReadMessage() (Message, error) {
	var msgType int

	if err := cslq.Decode(ctl.rw, "c", &msgType); err != nil {
		return nil, err
	}

	switch msgType {
	case mcClose:
		return CloseMessage{}, nil

	case mcQuery:
		var msg QueryMessage
		var err = cslq.Decode(ctl.rw, "v", &msg)
		return msg, err

	case mcDrop:
		var msg DropMessage
		var err = cslq.Decode(ctl.rw, "v", &msg)
		return msg, err
	}

	return nil, errors.New("unknown message type")
}

func (ctl *Control) WriteClose() error {
	return cslq.Encode(ctl.rw, "c", mcClose)
}

func (ctl *Control) WriteQuery(query string, port int) error {
	return cslq.Encode(ctl.rw, "cv", mcQuery, &QueryMessage{
		query: query,
		port:  port,
	})
}

func (ctl *Control) WriteDrop(port int) error {
	return cslq.Encode(ctl.rw, "cv", mcDrop, &DropMessage{port: port})
}
