package rpc

import (
	"github.com/cryptopunkscc/astrald/cslq/rpc"
	"github.com/cryptopunkscc/astrald/data"
	"io"
)

type Session struct {
	*rpc.Session[string]
}

var es rpc.ErrorSpace

var (
	// ErrUnavailable - requested data is not available
	ErrUnavailable = es.NewError(0x01, "unavailable")

	// ErrSeekUnsupported - seek operation is not supported
	ErrSeekUnsupported = es.NewError(0x02, "seek unsupported")

	// ErrSizeMismatch - size from data.ID and actual data size differ
	ErrSizeMismatch = es.NewError(0x03, "block size mismatch")

	// ErrInvalidOffset - offset is less than zero or more than the data length
	ErrInvalidOffset = es.NewError(0x04, "invalid offset")

	// ErrInvalidLength - length exceeds the remaining data or lenght is less than 1
	ErrInvalidLength = es.NewError(0x05, "invalid length")
)

func New(rw io.ReadWriter) Session {
	var s = rpc.NewSession[string](rw, es)
	s.ErrorType = "c"
	return Session{s}
}

func (s Session) RegisterSource(serviceName string) error {
	if err := s.Encode(MsgRegisterSource{Service: serviceName}); err != nil {
		return err
	}

	return s.DecodeErr()
}

func (s Session) Read(dataID data.ID, offset int, length int) (io.Reader, error) {
	err := s.Encode(MsgRead{
		DataID: dataID,
		Start:  int64(offset),
		Len:    int64(length),
	})
	if err != nil {
		return nil, err
	}
	err = s.DecodeErr()
	if err != nil {
		return nil, err
	}
	return s.Session.Transport(), nil
}
